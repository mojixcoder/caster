package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/mojixcoder/caster/internal/app"
	"github.com/mojixcoder/caster/internal/cache"
	"github.com/mojixcoder/caster/internal/cluster"
	"github.com/mojixcoder/kid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	tracesdk "go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type (
	GetResponse struct {
		Value any `json:"value"`
	}

	SetRequest struct {
		Key   string `json:"key"`
		Value any    `json:"value"`
	}
)

var (
	EmptyResponse = []byte("{\"message\":\"ok\"}\n")

	ErrInternal = kid.Map{"message": "internal error"}

	ErrNotFound = kid.Map{"message": "not found"}

	ErrKeyRequired = kid.Map{"message": "key is required."}

	ErrNoKey = errors.New("key is required")
)

// initHandlers initializes HTTP handlers.
func (s Server) initHandlers() {
	g := s.kid.Group("", NewTraceMiddleware())

	g.Get("/get", s.GetFromCache)
	g.Post("/set", s.SetToCache)
	g.Get("/flush", s.FlushCache)
}

func (s Server) mergeAddressAndPath(address, path string) string {
	address = strings.TrimRight(address, "/")
	return address + path
}

func getSpan(c *kid.Context, name string) (context.Context, tracesdk.Span) {
	extractedCtx, _ := c.Get("ctx")
	ctx := extractedCtx.(context.Context)

	ctx, span := otel.Tracer(app.App.Config.Tracer.Name).Start(ctx, name)
	return ctx, span
}

func injectReq(ctx context.Context, req *http.Request) *http.Request {
	req = req.WithContext(ctx)
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
	return req
}

// GetFromCache gets a key from cache.
func (s Server) GetFromCache(c *kid.Context) {
	ctx, span := getSpan(c, "get_from_cache")
	defer span.End()

	key := c.QueryParam("key")
	if key == "" {
		span.RecordError(ErrNoKey)
		c.JSON(http.StatusBadRequest, ErrKeyRequired)
		return
	}

	node := s.cluster.GetNodeFromKey(key)
	isLocal := node.IsLocal()

	span.SetAttributes(attribute.Bool("is_local", isLocal))

	switch isLocal {
	// Is local node.
	case true:
		val, err := s.cache.Get(key)
		if err != nil {
			if err == cache.ErrNotFound {
				c.JSON(http.StatusNotFound, ErrNotFound)
				span.SetAttributes(attribute.Bool("key_found", false))
				return
			}
			app.App.Logger.Error("error in getting key from cache", zap.Error(err))
			span.RecordError(err)
			span.SetStatus(codes.Error, "error in getting key from cache")
			c.JSON(http.StatusInternalServerError, ErrInternal)
			return
		}
		span.SetAttributes(attribute.Bool("key_found", true))

		res := GetResponse{Value: val}
		c.JSON(http.StatusOK, &res)

	// Is not local node.
	case false:
		app.App.Logger.Debug("getting key from another node", zap.String("node", node.Address()))

		req, _ := http.NewRequest(http.MethodGet, s.mergeAddressAndPath(node.Address(), "/get")+"?key="+key, nil)
		req = injectReq(ctx, req)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			app.App.Logger.Error(
				"error in calling a cluster member",
				zap.String("node", node.Address()),
				zap.String("path", "/get"),
				zap.Error(err),
			)
			span.RecordError(err)
			span.SetStatus(codes.Error, "error in calling a cluster member")
			c.JSON(http.StatusInternalServerError, ErrInternal)
			return
		}
		defer res.Body.Close()

		bytes, err := io.ReadAll(res.Body)
		if err != nil {
			app.App.Logger.Error(
				"error in reading response body",
				zap.String("node", node.Address()),
				zap.String("path", "/get"),
				zap.Error(err),
			)
			span.RecordError(err)
			span.SetStatus(codes.Error, "error in reading response body")
			c.JSON(http.StatusInternalServerError, ErrInternal)
			return
		}

		c.SetResponseHeader("Content-Type", res.Header.Get("Content-Type"))
		c.Byte(res.StatusCode, bytes)
	}
}

// SetToCache sets a key-value pair to cache.
func (s Server) SetToCache(c *kid.Context) {
	ctx, span := getSpan(c, "set_to_cache")
	defer span.End()

	var req SetRequest
	if err := c.ReadJSON(&req); err != nil {
		app.App.Logger.Error("error in reading request body", zap.Error(err))
		span.RecordError(err)
		c.JSON(http.StatusBadRequest, kid.Map{"message": err.Error()})
		return
	}

	if req.Key == "" {
		span.RecordError(ErrNoKey)
		c.JSON(http.StatusBadRequest, ErrKeyRequired)
		return
	}

	node := s.cluster.GetNodeFromKey(req.Key)
	isLocal := node.IsLocal()

	span.SetAttributes(attribute.Bool("is_local", isLocal))

	switch isLocal {
	// Is local node.
	case true:
		if err := s.cache.Set(req.Key, req.Value); err != nil {
			app.App.Logger.Error("error in setting key to the cache", zap.Error(err))
			span.RecordError(err)
			span.SetStatus(codes.Error, "error in setting key to the cache")
			c.JSON(http.StatusInternalServerError, ErrInternal)
			return
		}

		c.SetResponseHeader("Content-Type", "application/json")
		c.Byte(http.StatusOK, EmptyResponse)

	// Is not local node.
	case false:
		app.App.Logger.Debug("setting key to another node", zap.String("node", node.Address()))

		jsonBytes, _ := json.Marshal(&req)

		req, _ := http.NewRequest(http.MethodPost, s.mergeAddressAndPath(node.Address(), "/set"), bytes.NewBuffer(jsonBytes))
		req = injectReq(ctx, req)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			app.App.Logger.Error(
				"error in calling a cluster member",
				zap.String("node", node.Address()),
				zap.String("path", "/set"),
				zap.Error(err),
			)
			span.RecordError(err)
			span.SetStatus(codes.Error, "error in calling a cluster member")
			c.JSON(http.StatusInternalServerError, ErrInternal)
			return
		}
		defer res.Body.Close()

		bytes, err := io.ReadAll(res.Body)
		if err != nil {
			app.App.Logger.Error(
				"error in reading response body",
				zap.String("node", node.Address()),
				zap.String("path", "/set"),
				zap.Error(err),
			)
			span.RecordError(err)
			span.SetStatus(codes.Error, "error in reading response body")
			c.JSON(http.StatusInternalServerError, ErrInternal)
			return
		}

		c.SetResponseHeader("Content-Type", res.Header.Get("Content-Type"))
		c.Byte(res.StatusCode, bytes)
	}
}

// FlushCache clears cache.
func (s Server) FlushCache(c *kid.Context) {
	ctx, span := getSpan(c, "get_from_cache")
	defer span.End()

	flushAll, _ := strconv.ParseBool(c.QueryParam("all"))

	span.SetAttributes(attribute.Bool("flush_all", flushAll))

	if err := s.cache.Flush(); err != nil {
		app.App.Logger.Error("error in flushing cache", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrInternal)
		return
	}

	if flushAll {
		app.App.Logger.Debug("flushing other nodes")

		nodes := s.cluster.NonLocalNodes()
		var wg sync.WaitGroup
		var i int
		errs := make([]error, len(nodes))

		wg.Add(len(nodes))
		for _, node := range nodes {
			go func(i int, node cluster.Node) {
				defer wg.Done()

				req, _ := http.NewRequest(http.MethodGet, s.mergeAddressAndPath(node.Address(), "/flush?all=false"), nil)
				req = injectReq(ctx, req)

				_, err := http.DefaultClient.Do(req)
				errs[i] = err
			}(i, node)
			i++
		}
		wg.Wait()

		if err := errors.Join(errs...); err != nil {
			app.App.Logger.Error(
				"flushing the cache of some nodes failed",
				zap.Error(err),
				zap.String("path", "/flush"),
			)
			span.RecordError(err)
			span.SetStatus(codes.Error, "flushing the cache of some nodes failed")
			c.JSON(http.StatusInternalServerError, ErrInternal)
			return
		}
	}

	c.SetResponseHeader("Content-Type", "application/json")
	c.Byte(http.StatusOK, EmptyResponse)
}

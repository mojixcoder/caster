package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/mojixcoder/caster/internal/app"
	"github.com/mojixcoder/caster/internal/cache"
	"github.com/mojixcoder/kid"
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
	EmptyResponse = []byte("{\"message\":\"ok\"}")

	ErrInternal = kid.Map{"message": "internal error"}

	ErrNotFound = kid.Map{"message": "not found"}

	ErrKeyRequired = kid.Map{"message": "key is required."}
)

// initHandlers initializes HTTP handlers.
func (s Server) initHandlers() {
	s.kid.Get("/get", s.GetFromCache)
	s.kid.Post("/set", s.SetToCache)
}

func (s Server) mergeAddressAndPath(address, path string) string {
	address = strings.TrimRight(address, "/")
	return address + path
}

// GetFromCache gets a key from cache.
func (s Server) GetFromCache(c *kid.Context) {
	key := c.QueryParam("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, ErrKeyRequired)
		return
	}

	node := s.cluster.GetNodeFromKey(key)

	switch node.IsLocal() {
	// Is local node.
	case true:
		val, err := s.cache.Get(key)
		if err != nil {
			if err == cache.ErrNotFound {
				c.JSON(http.StatusNotFound, ErrNotFound)
				return
			}
			app.App.Logger.Error("error in getting key from cache", zap.Error(err))
			c.JSON(http.StatusInternalServerError, ErrInternal)
			return
		}

		res := GetResponse{Value: val}
		c.JSON(http.StatusOK, &res)

	// Is not local node.
	case false:
		app.App.Logger.Debug("getting key from another node", zap.String("node", node.Address()))

		res, err := http.Get(s.mergeAddressAndPath(node.Address(), "/get") + "?key=" + key)
		if err != nil {
			app.App.Logger.Error(
				"error in calling a cluster member",
				zap.String("node", node.Address()),
				zap.String("path", "/get"),
				zap.Error(err),
			)
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
			c.JSON(http.StatusInternalServerError, ErrInternal)
			return
		}

		c.SetResponseHeader("Content-Type", res.Header.Get("Content-Type"))
		c.Byte(res.StatusCode, bytes)
	}
}

// SetToCache sets a key-value pair to cache.
func (s Server) SetToCache(c *kid.Context) {
	var req SetRequest
	if err := c.ReadJSON(&req); err != nil {
		app.App.Logger.Error("error in reading request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, kid.Map{"message": err.Error()})
		return
	}

	if req.Key == "" {
		c.JSON(http.StatusBadRequest, ErrKeyRequired)
		return
	}

	node := s.cluster.GetNodeFromKey(req.Key)

	switch node.IsLocal() {
	// Is local node.
	case true:
		if err := s.cache.Set(req.Key, req.Value); err != nil {
			app.App.Logger.Error("error in setting key to the cache", zap.Error(err))
			c.JSON(http.StatusInternalServerError, ErrInternal)
			return
		}

		c.SetResponseHeader("Content-Type", "application/json")
		c.Byte(http.StatusOK, EmptyResponse)

	// Is not local node.
	case false:
		app.App.Logger.Debug("setting key to another node", zap.String("node", node.Address()))

		jsonBytes, _ := json.Marshal(&req)

		res, err := http.Post(
			s.mergeAddressAndPath(node.Address(), "/set"),
			"application/json",
			bytes.NewBuffer(jsonBytes),
		)
		if err != nil {
			app.App.Logger.Error(
				"error in calling a cluster member",
				zap.String("node", node.Address()),
				zap.String("path", "/set"),
				zap.Error(err),
			)
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
			c.JSON(http.StatusInternalServerError, ErrInternal)
			return
		}

		c.SetResponseHeader("Content-Type", res.Header.Get("Content-Type"))
		c.Byte(res.StatusCode, bytes)
	}
}

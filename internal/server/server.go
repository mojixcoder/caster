package server

import (
	"fmt"
	"net/http"

	"github.com/mojixcoder/caster/internal/app"
	"github.com/mojixcoder/caster/internal/cache"
	"github.com/mojixcoder/caster/internal/cluster"
	"github.com/mojixcoder/kid"
	"github.com/mojixcoder/kid/middlewares"
	"go.uber.org/zap"
)

// Server manages server-related stuff.
type Server struct {
	// cluster is the cluster manager.
	cluster cluster.Cluster

	// cache is the cache storage.
	cache cache.Cache

	kid *kid.Kid
}

// RunServer runs the server.
func (s *Server) RunServer() error {
	s.kid.ApplyOptions(kid.WithDebug(app.App.Config.Caster.Debug))

	s.kid.Use(middlewares.NewRecoveryWithConfig(middlewares.RecoveryConfig{
		OnRecovery: func(c *kid.Context, err any) {
			app.App.Logger.Error("panic recovered", zap.Any("reason", err))
			c.JSON(http.StatusInternalServerError, ErrInternal)
		},
	}))

	s.initHandlers()

	port := fmt.Sprintf(":%d", app.App.Config.Caster.Port)
	app.App.Logger.Info("running server", zap.String("address", "0.0.0.0"+port))

	return s.kid.Run(port)
}

// NewServer returns a new server.
func NewServer(cache cache.Cache, cluster cluster.Cluster) *Server {
	return &Server{cache: cache, cluster: cluster, kid: kid.New()}
}

package server

import (
	"fmt"

	"github.com/mojixcoder/caster/internal/app"
	"github.com/mojixcoder/caster/internal/cache"
	"github.com/mojixcoder/caster/internal/cluster"
	"github.com/mojixcoder/kid"
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
	s.initHandlers()

	port := fmt.Sprintf(":%d", app.App.Config.Caster.Port)
	app.App.Logger.Info("running server", zap.String("address", "0.0.0.0"+port))

	s.kid.ApplyOptions(kid.WithDebug(app.App.Config.Caster.Debug))

	return s.kid.Run(port)
}

// NewServer returns a new server.
func NewServer(cache cache.Cache, cluster cluster.Cluster) *Server {
	return &Server{cache: cache, cluster: cluster, kid: kid.New()}
}

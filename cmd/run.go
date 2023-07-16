package cmd

import (
	"fmt"
	"net/http"

	"github.com/mojixcoder/caster/internal/app"
	"github.com/mojixcoder/caster/internal/cache"
	"github.com/mojixcoder/caster/internal/cluster"
	"github.com/mojixcoder/caster/internal/server"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// runCmd is the root command.
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Caster is a cache database.",
	Run:   run,
}

func run(cmd *cobra.Command, args []string) {
	fmt.Println(app.Banner)

	app.Init()

	cache := cache.NewLRUCache()

	cluster, err := cluster.NewCluster()
	if err != nil {
		app.App.Logger.Fatal("error in creating the cluster", zap.Error(err))
	}

	srv := server.NewServer(cache, cluster)

	if err := srv.RunServer(); err != nil {
		if err != http.ErrServerClosed {
			app.App.Logger.Fatal("running server failed", zap.Error(err))
		}
	}

}

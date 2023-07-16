package app

import (
	"github.com/mojixcoder/caster/internal/config"
	"github.com/mojixcoder/caster/pkg/logger"
	"go.uber.org/zap"
)

var App *AppRepo

// Version is the app version.
var Version string

// Banner is the app Banner.
const Banner string = `
	██████╗ █████╗ ███████╗████████╗███████╗██████╗ 
	██╔════╝██╔══██╗██╔════╝╚══██╔══╝██╔════╝██╔══██╗
	██║     ███████║███████╗   ██║   █████╗  ██████╔╝
	██║     ██╔══██║╚════██║   ██║   ██╔══╝  ██╔══██╗
	╚██████╗██║  ██║███████║   ██║   ███████╗██║  ██║
	 ╚═════╝╚═╝  ╚═╝╚══════╝   ╚═╝   ╚══════╝╚═╝  ╚═╝`

// AppRepo holds things related to the entire application like config, logger, etc.
type AppRepo struct {
	// Logger is the logger used across the application.
	Logger *zap.Logger

	// Config is the application config.
	Config *config.AppConfig
}

// Init initializes the application related stuff.
func Init() {
	var app AppRepo

	cfg, err := config.Load()
	if err != nil {
		panic("error in creating logger, reason: " + err.Error())
	}

	logger, err := logger.New("caster", cfg.Caster.Debug)
	if err != nil {
		panic("error in creating logger, reason: " + err.Error())
	}

	app.Logger = logger
	app.Config = cfg

	App = &app
}

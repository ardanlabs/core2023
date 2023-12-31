package checkgrp

import (
	"net/http"

	"github.com/ardanlabs/service/foundation/logger"
	"github.com/ardanlabs/service/foundation/web"
	"github.com/jmoiron/sqlx"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Build string
	Log   *logger.Logger
	DB    *sqlx.DB
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	hdl := New(cfg.Build, cfg.Log, cfg.DB)
	app.Handle(http.MethodGet, "/readiness", hdl.Readiness)
	app.Handle(http.MethodGet, "/liveness", hdl.Liveness)
}

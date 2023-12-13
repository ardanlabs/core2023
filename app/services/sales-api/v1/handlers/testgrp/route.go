package testgrp

import (
	"net/http"

	"github.com/ardanlabs/service/business/web/v1/auth"
	"github.com/ardanlabs/service/business/web/v1/mid"
	"github.com/ardanlabs/service/business/web/v1/mux"
	"github.com/ardanlabs/service/foundation/web"
)

func Routes(app *web.App, cfg mux.Config) {
	authen := mid.Authenticate(cfg.Auth)
	ruleAdmin := mid.Authorize(cfg.Auth, auth.RuleAdminOnly)

	app.Handle(http.MethodGet, "/test", Test)
	app.Handle(http.MethodGet, "/auth", Test, authen, ruleAdmin)
}

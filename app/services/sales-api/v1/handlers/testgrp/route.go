package testgrp

import (
	"net/http"

	"github.com/ardanlabs/service/business/web/v1/mux"
	"github.com/ardanlabs/service/foundation/web"
)

func Route(app *web.App, cfg mux.Config) {
	app.Handle(http.MethodGet, "/test", Test)
}

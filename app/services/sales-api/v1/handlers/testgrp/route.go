package testgrp

import (
	"net/http"

	"github.com/ardanlabs/service/business/web/v1/mux"
	"github.com/dimfeld/httptreemux/v5"
)

func Route(mux *httptreemux.ContextMux, cfg mux.Config) {
	mux.Handle(http.MethodGet, "/test", Test)
}

package all

import (
	"github.com/ardanlabs/service/app/services/sales-api/v1/handlers/testgrp"
	"github.com/ardanlabs/service/business/web/v1/mux"
	"github.com/dimfeld/httptreemux/v5"
)

// Routes constructs the add value which provides the implementation of
// of RouteAdder for specifying what routes to bind to this instance.
func Routes() add {
	return add{}
}

type add struct{}

// Add implements the RouterAdder interface.
func (add) Add(mux *httptreemux.ContextMux, cfg mux.Config) {
	testgrp.Route(mux, cfg)
}

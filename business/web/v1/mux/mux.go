package mux

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/ardanlabs/service/foundation/logger"
	"github.com/dimfeld/httptreemux/v5"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Build    string
	Shutdown chan os.Signal
	Log      *logger.Logger
}

// WebAPI constructs a http.Handler with all application routes bound.
func WebAPI(cfg Config) *httptreemux.ContextMux {
	mux := httptreemux.NewContextMux()

	h := func(w http.ResponseWriter, r *http.Request) {
		status := struct {
			Status string
		}{
			Status: "OK",
		}
		json.NewEncoder(w).Encode(status)
	}

	mux.Handle(http.MethodGet, "/test", h)

	return mux
}

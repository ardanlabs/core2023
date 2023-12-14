package usergrp

import (
	"net/http"

	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/business/core/user/stores/userdb"
	"github.com/ardanlabs/service/business/web/v1/auth"
	"github.com/ardanlabs/service/business/web/v1/mid"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/ardanlabs/service/foundation/web"
	"github.com/jmoiron/sqlx"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log  *logger.Logger
	Auth *auth.Auth
	DB   *sqlx.DB
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	usrCore := user.NewCore(cfg.Log, userdb.NewStore(cfg.Log, cfg.DB))

	authen := mid.Authenticate(cfg.Auth)
	ruleAdmin := mid.Authorize(cfg.Auth, auth.RuleAdminOnly)
	ruleAdminOrSubject := mid.AuthorizeUser(cfg.Auth, auth.RuleAdminOrSubject, usrCore)

	hdl := New(usrCore, cfg.Auth)
	app.Handle(http.MethodGet, "/users/token/:kid", hdl.Token)
	app.Handle(http.MethodGet, "/users", hdl.Query, authen, ruleAdmin)
	app.Handle(http.MethodGet, "/users/:user_id", hdl.QueryByID, authen, ruleAdminOrSubject)
	app.Handle(http.MethodPost, "/users", hdl.Create, authen, ruleAdmin)
	app.Handle(http.MethodPut, "/users/:user_id", hdl.Update, authen, ruleAdminOrSubject)
	app.Handle(http.MethodDelete, "/users/:user_id", hdl.Delete, authen, ruleAdminOrSubject)
}

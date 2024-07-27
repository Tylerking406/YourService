package handlers

import (
	"dev/yourservice.git/business/yourservice"
	"dev/yourservice.git/business/i"
	"dev/yourservice.git/business/mid"
	"dev/yourservice.git/foundation/web"
	"net/http"
	"os"
)

type Yourservice struct {
	Service *yourservice.Service
}

// API constructs a http.Handler with all application routes defined
func API(log i.Logger, y Yourservice, shutdown chan os.Signal) *web.App {

	// Create web app with middleware
	app := web.NewApp(
		shutdown,
		mid.Logger(log),
		mid.Errors(log),
		mid.Panics(log),
	)

	// Check Service
	ch := check{}
	app.Handle(http.MethodGet, "/readiness", ch.readiness)
	app.Handle(http.MethodGet, "/liveliness", ch.liveliness)

	// Yourservice Handlers
	app.Handle(http.MethodPost, "/create", y.create)
	return app

}

// Init will initialise the Service
func Init(db yourservice.Store, log i.Logger) Yourservice {

	// Initialise services
	y := Yourservice{
		Service: &yourservice.Service{
			Log:   log,
			Store: db,
		},
	}
	return y

}

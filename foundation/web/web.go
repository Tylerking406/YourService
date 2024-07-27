// Package web contains a small web framework extension.
package web

import (
	"context"
	"github.com/dimfeld/httptreemux"
	"github.com/google/uuid"
	"net/http"
	"os"
	"syscall"
	"time"
)

// ctxKey represents the type of value for the context key.
type ctxKey int64

// KeyValues is how request values are stored/retrieved.
const KeyValues ctxKey = 1

// Values represent state for each request.
type Values struct {
	TraceID    string
	Now        time.Time
	StatusCode int
}

// A Handler is a type that handles an http request within our own little mini
// framework.
type Handler func(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
) error

// registered keeps track of handlers registered to the http default server
// mux. This is a singleton and used by the standard library for metrics
// and profiling. The application may want to add other handlers like
// readiness and liveness to that mux. If this is not tracked, the routes
// could try to be registered more than once, causing a panic.
var registered = make(map[string]bool)

// App is the entrypoint into our application and what configures our context
// object for each of our http handlers. Feel free to add any configuration
// data/logic on this App struct.
type App struct {
	mux      *httptreemux.ContextMux
	shutdown chan os.Signal
	mw       []Middleware
}

// IsDevAppServer will return true if we are running locally
// It the code is running in the cloud, then it will return false
func IsDevAppServer() bool {
	return os.Getenv("GAE_DEPLOYMENT_ID") == ""
}

// NewApp creates an App value that handle a set of routes for the application.
func NewApp(shutdown chan os.Signal, mw ...Middleware) *App {

	mux := httptreemux.NewContextMux()

	return &App{
		mux:      mux,
		shutdown: shutdown,
		mw:       mw,
	}
}

// SignalShutdown is used to gracefully shutdown the app when an integrity issue
// is identified.
func (a *App) SignalShutdown() {
	a.shutdown <- syscall.SIGTERM
}

// ServeHTTP implements the http.Handler interface. It's the entry point for all
// http traffic and allows the opentelemetry mux to run first to handle tracing.
// The opentelemetry mux then calls the application mux to handle application
// traffic.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mux.ServeHTTP(w, r)
}

// Handle sets a handler function for a given HTTP method and path pair to the
// application server mux.
func (a *App) Handle(
	method string,
	path string,
	handler Handler,
	mw ...Middleware,
) {
	a.handle(false, method, path, handler, mw...)
}

// handle performs the real work of applying boilerplate and framework code for
// a handler.
func (a *App) handle(
	debug bool,
	method string,
	path string,
	handler Handler,
	mw ...Middleware,
) {
	if debug {
		// Track all the handlers that are being registered so we don't have the
		// same handlers registered twice to this singleton.
		if _, exists := registered[method+path]; exists {
			return
		}
		registered[method+path] = true
	}

	// First wrap handler specific middleware around this handler.
	handler = wrapMiddleware(mw, handler)

	// Add the application's general middleware to the handler chain.
	handler = wrapMiddleware(a.mw, handler)

	// The function to execute for each request.
	h := func(w http.ResponseWriter, r *http.Request) {

		// Set the context with the required values to
		// process the request.
		v := Values{
			TraceID: uuid.New().String()[:8],
			Now:     time.Now(),
		}
		ctx := context.WithValue(r.Context(), KeyValues, &v)

		// Call the wrapped handler functions.
		if err := handler(ctx, w, r); err != nil {
			// If we get an error at this level we are way beyond handling it.
			// If an error gets this far we have to shutdown the app, this is
			// foundational code.
			a.SignalShutdown()
			return
		}
	}

	a.mux.Handle(method, path, h)

}

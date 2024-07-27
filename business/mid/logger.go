package mid

import (
	"context"
	"dev/yourservice.git/business/i"
	"dev/yourservice.git/foundation/web"
	"net/http"
	"time"
)

// Logger writes some information about the request to the logs in the
// format: TraceID : (200) GET /foo -> IP ADDR (latency)
func Logger(log i.Logger) web.Middleware {

	// This is the actual middleware function to be executed.
	m := func(handler web.Handler) web.Handler {

		// Create the handler that will be attached in the middleware chain.
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

			// If the context is missing this value, request the service
			// to be shutdown gracefully.
			v, ok := ctx.Value(web.KeyValues).(*web.Values)
			if !ok {
				return web.NewShutdownError("web value missing from context")
			}

			log.Printf("[%v]: started   : [%v] [%v] -> [%v]",
				v.TraceID,
				r.Method, r.URL.Path, r.RemoteAddr,
			)

			// Call the next handler.
			err := handler(ctx, w, r)

			log.Printf("[%v]: completed : [%v] [%v] -> [%v] ([%v]) ([%v])",
				v.TraceID,
				r.Method, r.URL.Path, r.RemoteAddr,
				v.StatusCode, time.Since(v.Now),
			)

			// Return the error so it can be handled further up the chain.
			return err
		}

		return h
	}

	return m
}

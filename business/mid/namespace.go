package mid

import (
	"context"
	"dev/yourservice.git/foundation/web"
	"net/http"
)

// Namespace will validate the namespace, ensuring it is not empty or the default identifier & then sets it on the context
// Takes in an optional namespace parameter:
// If supplied, the namespace from the endpoint request will be compared against the namespace parameter to be equal
// If the namespace from the endpoint request is empty, namespace will be set to the namespace parameter
func Namespace(namespace ...string) web.Middleware {

	// This is the actual middleware function to be executed.
	m := func(handler web.Handler) web.Handler {

		// Create the handler that will be attached in the middleware chain.
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) (err error) {

			// If the context is missing this value, request the service to be shutdown gracefully
			_, ok := ctx.Value(web.KeyValues).(*web.Values)
			if !ok {
				return web.NewShutdownError("web value missing from context")
			}

			// Get namespace from url parameter
			ns := web.GetParam(r, "ns")
			if (namespace[0] == "" && ns == "") || ns == "__$DEFAULT$__" {
				return web.Errorf("namespace cannot be empty or the default identifier")
			}

			// Overwrite namespace with application defined namespace
			if namespace[0] != "" {
				if ns != "" && ns != namespace[0] {
					return web.Errorf("namespace must be [%v], but got [%v]", namespace[0], ns)
				}
				ns = namespace[0]
			}

			// Add namespace to ctx
			ctx = context.WithValue(ctx, "namespace", ns)

			// Call the next handler and set its return value in the err variable.
			return handler(ctx, w, r)
		}

		return h
	}

	return m
}

package handlers

import (
	"context"
	"dev/yourservice.git/foundation/web"
	"net/http"
)

// create ...
func (y Yourservice) create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	// Get request data
	var request = struct {
		Value string `validate:"required"`
	}{}

	// Decode, sanitize & validate request
	err := web.Decode(r, &request)
	if err != nil {
		return err
	}

	// Log
	y.Service.Log.Printf("Creating...")

	// Create
	err = y.Service.Create(ctx)
	if err != nil {
		return err
	}

	// Send response data
	response := struct {
		Status string `json:"Status"`
	}{
		Status: "Success",
	}
	return web.Respond(ctx, w, response, http.StatusOK)

}

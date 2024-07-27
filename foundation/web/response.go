package web

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"log"
	"net/http"
)

// Respond converts a Go value to JSON and sends it to the client.
// data must be a *S
func Respond(
	ctx context.Context,
	w http.ResponseWriter,
	data interface{},
	statusCode int,
) error {

	// Set the status code for the request logger middleware. If the context is
	// missing this value, request the service to be shutdown gracefully.
	v, ok := ctx.Value(KeyValues).(*Values)
	if !ok {
		return NewShutdownError("web value missing from context")
	}
	v.StatusCode = statusCode

	// If there is nothing to marshal then set status code and return.
	if statusCode == http.StatusNoContent {
		w.WriteHeader(statusCode)
		return nil
	}

	// Convert the response value to JSON.
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Set the content type and headers once we know marshaling has succeeded.
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, OPTIONS, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "content-type, authorization, disbursetotalcount, disbursetotalsum, uidx")

	// Write the status code to the response.
	w.WriteHeader(statusCode)
	resp := string(jsonData)
	log.Print(resp)

	// Send the result back to the client.
	_, err = w.Write(jsonData)
	if err != nil {
		return err
	}
	return nil

}

// RespondError sends an error reponse back to the client. Should not be called
// by a handler functions unless for specific business reasons. Generally,
// errors should be handled by middleware.
func RespondError(
	ctx context.Context,
	w http.ResponseWriter,
	err error,
) error {

	// If the error was of the type *Error, the handler has
	// a specific status code and error to return.
	webErr, ok := errors.Cause(err).(*Error)
	if ok {
		er := ErrorResponse{
			Error:  webErr.Err.Error(),
			Fields: webErr.Fields,
		}
		if err := Respond(ctx, w, er, webErr.StatusCode); err != nil {
			return err
		}
		return nil
	}

	// If not, the handler sent any arbitrary error value so use 500.
	err_ := ErrorResponse{
		Error: http.StatusText(http.StatusInternalServerError),
	}
	err = Respond(ctx, w, err_, http.StatusInternalServerError)
	if err != nil {
		return err
	}
	return nil

}

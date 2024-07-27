package handlers

import (
	"context"
	"dev/yourservice.git/foundation/web"
	"net/http"
	"os"
)

type check struct{}

// readiness simply returns a 200 ok when called
func (c check) readiness(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	status := struct{ Status string }{
		Status: "OK",
	}
	return web.Respond(ctx, w, status, http.StatusOK)

}

// liveliness returns simple status info if the service is alive
func (c check) liveliness(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	host, err := os.Hostname()
	if err != nil {
		host = "hostname unavailable"
	}
	info := struct {
		Status             string `json:"Status,omitempty"`
		Build              string `json:"Build,omitempty"`
		Host               string `json:"Host,omitempty"`
		Application        string `json:"Application,omitempty"`
		DeploymentID       string `json:"DeploymentID,omitempty"`
		Env                string `json:"Env,omitempty"`
		Instance           string `json:"Instance,omitempty"`
		MemoryMB           string `json:"MemoryMB,omitempty"`
		Runtime            string `json:"Runtime,omitempty"`
		Service            string `json:"Service,omitempty"`
		Version            string `json:"Version,omitempty"`
		GoogleCloudProject string `json:"GoogleCloudProject,omitempty"`
		Port               string `json:"Port,omitempty"`
	}{
		Status:             "up",
		Host:               host,
		Application:        os.Getenv("GAE_APPLICATION"),
		DeploymentID:       os.Getenv("GAE_DEPLOYMENT_ID"),
		Env:                os.Getenv("GAE_ENV"),
		Instance:           os.Getenv("GAE_INSTANCE"),
		MemoryMB:           os.Getenv("GAE_MEMORY_MB"),
		Runtime:            os.Getenv("GAE_RUNTIME"),
		Service:            os.Getenv("GAE_SERVICE"),
		Version:            os.Getenv("GAE_VERSION"),
		GoogleCloudProject: os.Getenv("GOOGLE_CLOUD_PROJECT"),
		Port:               os.Getenv("PORT"),
	}
	return web.Respond(ctx, w, info, http.StatusOK)

}

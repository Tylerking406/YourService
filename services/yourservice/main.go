package main

import (
	"context"
	"dev/yourservice.git/services/yourservice/handlers"
	some_db "dev/yourservice.git/thirdparty/some-db"
	"fmt"
	"github.com/ardanlabs/conf/v2"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	// Call run to wrap error
	err := run(log.New(os.Stdout, "", 0))
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

}

func run(log *log.Logger) error {

	// Configuration uses github.com/ardanlabs/conf/v2 library
	// Your program configuration is attempted to be retrieved in the priority:
	// 1) Environment variable
	// 2) CMD flag
	// 3) Else the default value will be used
	defer log.Println("Completed")
	var cfg struct {
		Web struct {
			APIHost         string        `conf:"default:0.0.0.0:8080"`
			ReadTimeout     time.Duration `conf:"default:5s"`
			ShutdownTimeout time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:0s"`
		}
	}
	namespace := "YOURSERVICE"
	data, err := conf.Parse(namespace, &cfg)
	if errors.Is(err, conf.ErrHelpWanted) {
		if data == "" {
			fmt.Println("version is not set")
		}
		fmt.Println(data)
		return nil
	}
	out, err := conf.String(&cfg)
	if err != nil {
		return err
	}
	log.Printf("Config:\n%v\n", out)

	// Initialise dependencies for later dependency injection
	log.Println("Initialising Services")
	db, err := some_db.NewClient(log)
	if err != nil {
		return err
	}
	defer func() {
		log.Println("Shutting down services")
		db.Close()
	}()

	// Initialise YourService Service
	yourservice := handlers.Init(db, log)

	// Make a channel to listen for errors coming from the listener
	serverErrors := make(chan error, 1)

	// Make a channel to listen for an interrupt or terminate signal from the OS
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Initialise web app
	webApp := handlers.API(log, yourservice, shutdown)

	// Create the server that will listen and serve
	api := http.Server{
		Addr:         cfg.Web.APIHost,
		Handler:      webApp,
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
	}

	// Start the service listening for requests
	go func() {
		log.Printf("API listening on [%v]", api.Addr)
		serverErrors <- api.ListenAndServe()
	}()

	// Blocking main and waiting for shutdown
	select {
	case err := <-serverErrors:
		return errors.Wrap(err, "server error")
	case sig := <-shutdown:
		log.Printf("[%v] : Start shutdown", sig)

		// Give outstanding requests a deadline for completion
		ctx, cancel := context.WithTimeout(
			context.Background(),
			cfg.Web.ShutdownTimeout,
		)
		defer cancel()

		// Asking listener to shut down and shed load
		err = api.Shutdown(ctx)
		if err != nil {
			_ = api.Close()
			return errors.Wrap(err, "could not stop server gracefully")
		}
	}
	return nil

}

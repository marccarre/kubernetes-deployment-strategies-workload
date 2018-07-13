package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/gorilla/mux"         // Better HTTP API.
	log "github.com/sirupsen/logrus" // Better Logging.
	flag "github.com/spf13/pflag"    // POSIX/GNU-style CLI arguments.

	"github.com/marccarre/kubernetes-deployment-strategies-workload/pkg/db"
	"github.com/marccarre/kubernetes-deployment-strategies-workload/pkg/server"
)

func main() {
	dbConfig, httpConfig := parseCLIArguments()

	// Gracefully shut down on SIGINT (ctrl+c):
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Create the database client:
	db, err := db.NewPostgreSQLDB(dbConfig)
	if err != nil {
		log.WithField("err", err).Fatal("failed to create database client")
	}

	// Create the HTTP server:
	httpServer := newHTTPServer(httpConfig, db)

	// Run the server in a goroutine so that it doesn't block:
	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			log.WithField("err", err).Error("HTTP server stopped unexpectedly")
		}
	}()

	// Block until we receive the signal to quit:
	<-stop

	log.Info("shutting down...")
	httpServer.Shutdown(context.Background())
	log.Info("bye!")
	os.Exit(0)
}

func parseCLIArguments() (*db.Config, *server.Config) {
	// Parse CLI arguments into config object:
	dbConfig := &db.Config{}
	dbConfig.RegisterFlags(flag.CommandLine)
	httpConfig := &server.Config{}
	httpConfig.RegisterFlags(flag.CommandLine)
	flag.Parse()
	return dbConfig, httpConfig
}

func newHTTPServer(httpConfig *server.Config, db db.DB) *http.Server {
	server := server.New(db)
	router := mux.NewRouter()
	server.RegisterRoutes(router)
	return &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%v", httpConfig.Port),
		Handler: router,
		// Good practice to set timeouts to avoid Slowloris attacks.
		ReadTimeout:  httpConfig.ReadTimeout,
		WriteTimeout: httpConfig.WriteTimeout,
		IdleTimeout:  httpConfig.IdleTimeout,
	}
}

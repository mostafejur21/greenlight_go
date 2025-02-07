package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve() error {
	// Declare a HTTP server using the same settings as in our main() function
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(app.logger.Handler(), slog.LevelError),
	}

	// Start a background go routine
	go func() {
		// Create a quick channel which carries os.Single values
		quit := make(chan os.Signal, 1)

        // Use signal.Notify() to listen for incoming SIGINT and SIGTERM signals and
        // relay them to the quit channel. Any other signals will not be caught by
        // signal.Notify() and will retain their default behaviour
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

        // Read the signal from the quit channel. This code will block until a signal is received
		s := <-quit
		app.logger.Info("caught signal", "signal", s.String())
		os.Exit(0)
	}()

	// log a starting server message
	app.logger.Info("starting server", "addr", srv.Addr, "env", app.config.env)

	// start the server as normal, return any error
	return srv.ListenAndServe()
}

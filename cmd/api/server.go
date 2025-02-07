package main

import (
	"context"
	"errors"
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

    // Creating a shutdownError channel to return any error returned by the Shutdown() function
    shutdownError := make(chan error)

	// Start a background go routine
	go func() {
		// Create a quick channel which carries os.Single values
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		app.logger.Info("Shutting down server", "signal", s.String())

        // Create a context with a 30 second timeout.
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()

        // Call Shutdown() function on our server, passing the context we just made.
        // Shutdown() will return nil if the graceful shutdown was successful, or an
        // error (which may happen because of a problem closing the listeners, or
        // because the shutdown didn't complete before the context 30 second deadline).
        shutdownError <-srv.Shutdown(ctx)
	}()

	// log a starting server message
	app.logger.Info("starting server", "addr", srv.Addr, "env", app.config.env)


	// start the server as normal, return any error
    err := srv.ListenAndServe()
    if !errors.Is(err, http.ErrServerClosed) {
        return err
    }


    // otherwise, we wait to reveive the return value from the Shutdown() on the error channel.
    err = <-shutdownError
    if err != nil{
        return err
    }


    // at this point the shutdown is complete, so showing a message
    app.logger.Info("stopped server", "addr", srv.Addr)
    return nil
}

package main

import (
	"context"
	"net/http"

	"github.com/mostafejur21/greenlight_go/internal/data"
)

// Custom context key type string
type contextKey string

const userContextKey = contextKey("user")

// This method will return a new copy of the request with the provided User struct added
// to the context
func (app *application) contextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}


// This will return User struct from the request context
func (app *application) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user value in the request context")
	}

	return user
}

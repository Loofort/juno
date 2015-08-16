package middle

import (
	"encoding/base64"
	"golang.org/x/net/context"
	"juno/common"
	"juno/model/storage"
	"net/http"
)

// authMW is the Authentication aware router type
type authMW struct {
	base    ContextRouter
	storage storage.Storage
}

// Authentication returns router that perform Authentication check before handle requests
func Authentication(base ContextRouter, storage storage.Storage) ContextRouter {
	return authMW{base, storage}
}

// Handle add authorization check middleware before handler call.
// It stores auth info in context
func (mw authMW) Handle(method, path string, handler JunoHandler) {
	authHandler := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		const basicPrefix string = "Basic "

		// Get the Basic Authentication credentials
		auth := r.Header.Get("Authorization")
		if strings.HasPrefix(auth, basicPrefix) {
			// Check credentials
			payload, err := base64.StdEncoding.DecodeString(auth[len(basicPrefix):])
			if err == nil {
				pair := bytes.SplitN(payload, []byte(":"), 2)
				if len(pair) == 2 {
					// check user in storage.
					user, ok := mw.storage.UserByCreds(string(pair[0]), string(pair[1]))
					if !ok {
						common.SendErr(w, http.StatusForbidden, "Forbidden")
					}

					// put user to context
					setCtxUser(ctx, user)

					// Delegate request to the given handle
					handler(ctx, w, r)
					return
				}
			}
		}

		// Request Basic Authentication otherwise
		w.Header().Set("WWW-Authenticate", "Basic realm=\"Private Area\"")
		common.SendErr(w, http.StatusUnauthorized, "Unauthorized")
	}

	// configure base router with auth handler
	mw.base.Handle(method, path, authHandler)
}

// add user object to conext
func setCtxUser(ctx context.Context, user model.User) context.Context {
	return context.WithValue(ctx, userKey, p)
}

// CtxUser returns User saved in context, second variable is false when no user is found (anonymous mode)
func CtxUser(ctx context.Context, name string) (string, bool) {
	user, ok := ctx.Value(userKey).(model.User)
	return user, ok
}

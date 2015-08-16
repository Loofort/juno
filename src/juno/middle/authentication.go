package middle

import (
	"encoding/base64"
	"golang.org/x/net/context"
	"juno/common"
	"juno/model"
	"juno/model/storage"
	"net/http"
)

// authMW is the Authentication aware router type
type authMW struct {
	base ContextRouter
	stg  storage.Storage
}

// Authentication returns router that perform Authentication check before handle requests
func Authentication(base ContextRouter, stg storage.Storage) ContextRouter {
	return authMW{base, stg}
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
					user, ok := mw.stg.UserByCreds(string(pair[0]), string(pair[1]))
					if !ok {
						common.SendErr(w, http.StatusForbidden, "Forbidden")
					}

					// put user to context
					ctx = setCtxUser(ctx, user)

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

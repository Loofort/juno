package middle

import (
	"github.com/dimfeld/httptreemux"
	"golang.org/x/net/context"
	"juno/model"
	"juno/model/storage"
	"net/http"
)

// following types are common for each router middlewares
type (
	// Router defines only one method of base router that we want to overwrite in custom implementations
	Router interface {
		Handle(method, path string, handler httptreemux.HandlerFunc)
	}
	// ContextRouter operates with context handlers
	ContextRouter interface {
		Handle(method, path string, handler JunoHandler)
	}
	// JunoHandler is type of context aware handlers that are controller handlers
	JunoHandler func(context.Context, http.ResponseWriter, *http.Request)
)

// contextMW implements ContextRouter interface
type contextMW struct {
	base Router
	stg  storage.Storage
}

// Context creates context aware wrapper for usual router
func Context(base Router, stg storage.Storage) ContextRouter {
	return contextMW{base, stg}
}

// Handle creates adapter handler to be called by usual router.
// Adapter handler creates context and passes it to final handler
func (mw contextMW) Handle(method, path string, handler JunoHandler) {
	adapter := func(w http.ResponseWriter, r *http.Request, p map[string]string) {
		// context is used to preserve auth info (used by storage layer to check access permissions).
		// it also might be used to cancel current session, by request or by timeout (not implemented).

		// there is no timeout requirements, so create just cancelable context
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// add Params to context
		ctx = setCtxParam(ctx, p)

		// add anonym user. It might be overrided in authentication mw
		ctx = model.SetCtxUser(ctx, model.Anonym())

		// storage may want to reserve resource per session that has to release at the end
		// put it in context too
		ctx, release := mw.stg.Reserve(ctx)
		defer release()

		handler(ctx, w, r)
	}
	mw.base.Handle(method, path, adapter)
}

// To avoid key collisions in context we defines an unexported type key
type ctxKey int

var paramsKey ctxKey = 0

// setCtxParams adds params to context
func setCtxParam(ctx context.Context, p map[string]string) context.Context {
	return context.WithValue(ctx, paramsKey, p)
}

// CtxParams obtains param by name from context , second variable is false when no param is found
func CtxParam(ctx context.Context, name string) (string, bool) {
	params, ok := ctx.Value(paramsKey).(map[string]string)
	if !ok {
		return "", ok
	}
	p, ok := params[name]
	return p, ok
}

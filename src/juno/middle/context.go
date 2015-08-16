package middle

import (
	"github.com/julienschmidt/httprouter"
	"golang.org/x/net/context"
)

// following types are common for each router middlewares
type (
	// Router defines only one method of httprouter that we want to overwrite in custom implementations
	Router interface {
		Handle(method, path string, handler httprouter.Handle)
	}
	// ContextRouter operates with context handlers
	ContextRouter interface {
		Handle(method, path string, handler Handle)
	}
	// Handle is type of context aware handlers that are controller handlers
	Handle func(context.Context, http.ResponseWriter, *http.Request)
)

// contextMW implements ContextRouter interface
type contextMW struct {
	base Router
}

// Context creates context aware wrapper for usual router
func Context(base Router) ContextRouter {
	return contextMW{base}
}

// Handle creates adapter handler to be called by usual router.
// Adapter handler creates context and passes it to final handler
func (mw contextMW) Handle(method, path string, handler Handle) {
	adapter := func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

		// there is no timeout requirements, so create just cancelable context
		ctx, cancel = context.WithCancel(context.Background())
		defer cancel()

		// add httprouter.Params to context
		ctx = setCtxParams(ctx, ps)

		handler(ctx, w, r)
	}
	mw.base.Handle(method, path, adapter)
}

// To avoid key collisions in context we defines an unexported type key
type ctxKey int

var paramsKey ctxKey = 0

// setCtxParams adds httprouter params to context
func setCtxParams(ctx context.Context, ps httprouter.Params) context.Context {
	return context.WithValue(ctx, paramsKey, ps)
}

// CtxParams obtains httprouter params from context , second variable is false when no params is found
func CtxParams(ctx context.Context, ps httprouter.Params) (httprouter.Params, bool) {
	return ctx.Value(userIPKey).(httprouter.Params)
}

package middle

import (
	"github.com/dimfeld/httptreemux"
	"golang.org/x/net/context"
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
}

// Context creates context aware wrapper for usual router
func Context(base Router) ContextRouter {
	return contextMW{base}
}

// Handle creates adapter handler to be called by usual router.
// Adapter handler creates context and passes it to final handler
func (mw contextMW) Handle(method, path string, handler JunoHandler) {
	adapter := func(w http.ResponseWriter, r *http.Request, p map[string]string) {

		// there is no timeout requirements, so create just cancelable context
		ctx, cancel = context.WithCancel(context.Background())
		defer cancel()

		// add Params to context
		ctx = setCtxParams(ctx, p)

		handler(ctx, w, r)
	}
	mw.base.Handle(method, path, adapter)
}

// To avoid key collisions in context we defines an unexported type key
type ctxKey int

var paramsKey ctxKey = 0
var userKey ctxKey = 1

// setCtxParams adds params to context
func setCtxParams(ctx context.Context, p map[string]string) context.Context {
	return context.WithValue(ctx, paramsKey, p)
}

// CtxParams obtains param by name from context , second variable is false when no param is found
func CtxParams(ctx context.Context, name string) (string, bool) {
	p, ok := ctx.Value(paramsKey).(map[string]string)
	if !ok {
		return "", ok
	}
	return p[name]
}

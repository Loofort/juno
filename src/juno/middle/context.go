package middle

import (
	"github.com/dimfeld/httptreemux"
	"golang.org/x/net/context"
	"log"
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
		ctx = setCtxParam(ctx, p)

		// add anonym user. It might be overrided in authentication mw
		ctx = setCtxUser(ctx, model.Anonym())

		handler(ctx, w, r)
	}
	mw.base.Handle(method, path, adapter)
}

// To avoid key collisions in context we defines an unexported type key
type ctxKey int

var paramsKey ctxKey = 0
var userKey ctxKey = 1

// setCtxParams adds params to context
func setCtxParam(ctx context.Context, p map[string]string) context.Context {
	return context.WithValue(ctx, paramsKey, p)
}

// CtxParams obtains param by name from context , second variable is false when no param is found
func CtxParam(ctx context.Context, name string) (string, bool) {
	p, ok := ctx.Value(paramsKey).(map[string]string)
	if !ok {
		return "", ok
	}
	return p[name]
}

// setCtxUser adds user object to conext
func setCtxUser(ctx context.Context, user model.User) context.Context {
	return context.WithValue(ctx, userKey, p)
}

// CtxUser returns User from context
func CtxUser(ctx context.Context) model.User {
	user, ok := ctx.Value(userKey).(model.User)
	if !ok {
		log.Println("no user in context") // todo: write call stack
		user = model.Anonym()
	}
	return user
}

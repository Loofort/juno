package middle

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

// ForwardMW is the wrapper type for httprouter.Router that helps route collision pathes
type forwardMW struct {
	// we have to keep *httprouter.Router type instead Router type
	// becouse we use specific method .Lookup()
	base *httprouter.Router

	// if no specific route is found process will go to common handler
	common httprouter.Handle

	// set of specific Handlers
	specific map[string]httprouter.Handle
}

// Forward create wrapper for httprouter.Router to help route collision pathes
// This is very simple warapper, for every collision case use new instance
// NOTE: for now more common path (with :name notation) should be set first than more specific
func Forward(base *httprouter.Router) Router {
	return forwardMW{base}
}

// Handle add middleware before handler execution
// if path already matched to some route than the path will be processed in the middleware
func (mw forwardMW) Handle(method, path string, handler httprouter.Handle) {

	hdl, _, ok := mw.base.Lookup("GET", path)
	if hdl == nil && !ok {
		if mw.common != nil {
			panic(fmt.Sprintf("forwarder already is set, please use new instance for %s", path))
		}

		// no handler was assigned for this path
		// assign middleware to be able control collisions
		mw.base.Handle(method, path, mw.middleHandler)
		mw.common = handler
	} else {
		specific[path] = handler
	}

}

func (mw forwardMW) middleHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

}

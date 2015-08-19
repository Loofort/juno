package middle

import (
	"net/http"
	"strings"
)

//jsonChecker represents wrappe type
type jsonChecker struct {
	http.Handler
}

// JSONContentType wraps http.handler
// the Wrapper checks that request Content-Type is equal to "application/json"
func JSONContentType(h http.Handler) http.Handler {
	return jsonChecker{h}
}

// Overwrited method that performs the Content-Type check
func (ch jsonChecker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ct := r.Header.Get("Content-Type")
	switch r.Method {
	case "POST", "PUT", "PATCH":
		if strings.Index(ct, "application/json") == -1 {
			// send error as plain text
			http.Error(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
			return
		}
	}

	ch.Handler.ServeHTTP(w, r)
}

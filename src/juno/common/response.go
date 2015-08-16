package common

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// httpErr represent JSON error response
type httpErr struct {
	code    int
	message string
}

// SendErr responds to client with HTTP code and JSON error body
func SendErr(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	sned(w, httpErr{code, msg})
}

// marshal object to json and send to net
func send(w http.ResponseWriter, obj interface{}) {

	b, err := json.Marshal(obj)
	if err != nil {
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	n, err = w.Write(b)
	if err != nil {
		log.Println(err)
		return
	}
	if n != len(b) {
		log.Printf("sent output butes count %d, but expected %d", n, len(b))
		return
	}
}

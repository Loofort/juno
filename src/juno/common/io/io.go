package io

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// error messages displayed to client
const (
	ERR_DB        = "Oops! database problem, try again latter"
	ERR_NOPROF    = "profile not found"
	ERR_NOUSER    = "user not found"
	ERR_REQ       = "something wrong with your request body"
	ERR_FORBIDDEN = "Forbidden"
)

// Input obtains request json body and fills up object with data
func Input(r *http.Request, obj interface{}) error {
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(obj); err != nil {
		// todo: r.Body also should be written to log, but it needs to implement some protections
		// todo: json throws UnmarshalTypeError. we can handle it to determine bad field name and value to show to user.
		log.Println(err)
		return err
	}
	return nil
}

// Marshal object to json and send to net
// If http code isn't set it will be set to 200
func Output(w http.ResponseWriter, obj interface{}) {

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

// httpErr represent JSON error response
type ErrJSON struct {
	Code    int
	Message string
}

// Send error with code 400
func ErrClient(w http.ResponseWriter, msg string) {
	OutErr(w, http.StatusBadRequest, msg)
}

// Send error with code 500
func ErrServer(w http.ResponseWriter, msg string) {
	OutErr(w, http.StatusBadRequest, msg)
}

// Err responds to client with HTTP code and JSON error body
func Err(w http.ResponseWriter, msg string, code int) {
	w.WriteHeader(code)
	Output(w, ErrJSON{code, msg})
}

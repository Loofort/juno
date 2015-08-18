package main

import (
	"github.com/dimfeld/httptreemux"
	"juno/model/storage"
	"juno/server"
	"log"
	"net/http"
	"os"
)

func main() {
	// get config var
	port := envMustGet("JUNO_PORT")
	mhost := envMustGet("JUNO_MONGO_HOST")

	// initialize mongo
	s := storage.MgoMustConnect(mhost)
	defer s.Close()

	// controller have to work with storage
	c = controller.New(s)

	// init router. httptreemux is fast and convinient
	r := httptreemux.New()

	// some of the endpoints are available in anonymous mode and some of them aren't.
	// we can perform different middleware operations - with auth checks and without.

	// Context middleware creates request context and pass it to handlers
	rc := middle.Context(r)
	rc.Handle("POST", "/user", c.UserCreate)
	rc.Handle("GET", "/user/:userid/confirm", c.UserConfirm)

	rc.Handle("GET", "/profile/:profid", c.ProfileGet)
	rc.Handle("GET", "/profile/all", c.ProfileAll)

	// Add middleware that checks authentication.
	ra := middle.Authentication(rc, s)
	ra.Handle("PUT", "/profile", c.ProfileUpdate)
	ra.Handle("GET", "/profile/:profid/history", c.ProfileHistory)

	// add middleware that decorates router and checks that Content-Type is application/json
	r = middle.JSONContentType(r)

	// Fire up the server
	log.Panic(http.ListenAndServe("localhost:"+port, r))
}

// all env variable are necessary.
// It panics if var is missing, and program will stop.
func envMustGet(name string) string {
	str := os.Getenv(name)
	if env == "" {
		log.Panicf("can't find environment variable %s", name)
	}

	return str
}

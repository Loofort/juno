package main

import (
	"github.com/dimfeld/httptreemux"
	"juno/controller"
	"juno/middle"
	"juno/model/storage"
	"log"
	"net/http"
	"os"
)

const VER = "v1"

func main() {
	// get config var
	port := envMustGet("JUNO_PORT")
	murl := envMustGet("JUNO_MONGO_URL")

	// initialize mongo
	s := storage.MgoMustConnect(murl)
	defer s.Close()

	// controller have to work with storage
	c := controller.New(s)

	// init router. httptreemux is fast and convinient
	r := httptreemux.New()

	// some of the endpoints are available in anonymous mode and some of them aren't.
	// we can perform different middleware operations - with auth checks and without.

	// build middleware that creates context and pass it to handlers
	rc := middle.Context(r, s)
	// add version
	rc = middle.Version(rc, VER)
	rc.Handle("POST", "/user", c.UserCreate)
	rc.Handle("GET", "/user/:userid/confirm", c.UserConfirm)

	rc.Handle("GET", "/profile/:profid", c.ProfileGet)
	rc.Handle("GET", "/profile/all", c.ProfileAll)

	// Add middleware that checks authentication.
	ra := middle.Authentication(rc, s)
	ra.Handle("PUT", "/profile", c.ProfileUpdate)
	ra.Handle("GET", "/profile/:profid/history", c.ProfileHistory)

	// add middleware that decorates router and checks that Content-Type is application/json
	rj := middle.JSONContentType(r)

	// Fire up the server
	log.Panic(http.ListenAndServe("localhost:"+port, rj))
}

// several env variables are required.
// It panics if var is missing, and program will stop.
func envMustGet(name string) string {
	str := os.Getenv(name)
	if str == "" {
		log.Panicf("can't find environment variable %s", name)
	}

	return str
}

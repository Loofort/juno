package main

import (
	"github.com/julienschmidt/httprouter"
	"juno/model/storage"
	"juno/server"
	"log"
	"net/http"
	"os"
)

func main() {
	// get config var
	port := env("JUNO_PORT")
	mongoHost := env("JUNO_MONGO_HOST")

	// initialize mongo
	s := storage.New(mongoHost)
	defer s.Close()

	// controller have to work with storage
	c = controller.New(s)

	// some of the endpoints are available in anonymous mode and some of them aren't.
	// we can perform different middleware operations - with auth checks and without.

	// init router
	r := httprouter.New()

	// Context middleware creates request context and pass it to handlers
	rc := middle.Context(r)
	rc.Handle("POST", "/user", c.UserCreate)
	rc.Handle("GET", "/user/:userid/confirm", c.UserConfirm)

	// Add middleware that checks authentication
	ra := middle.Authentication(rc)
	ra.Handle("PUT", "/profile", c.ProfileUpdate)
	ra.Handle("GET", "/profile/:profid/history", c.ProfileHistory)

	// in following specific case httprouter can't differ urls like /:profid and /all
	// we need to handle it manually by our custom middleware forwarder.
	// add custom forwarder middleware and context
	f := middle.Forward(r)
	rc = middle.Context(f)
	rc.Handle("GET", "/profile/:profid", c.ProfileGet)
	rc.Handle("GET", "/profile/all", c.ProfileAll)

	// add middleware that decorates router and checks that Content-Type is application/json
	r = middle.JSONContentType(r)

	// Fire up the server
	log.Panic(http.ListenAndServe("localhost:"+port, r))

	// initialize router
	// user permissions are preserved in storage too
	ar := server.NewAuthRouter(s)

	// assign handlers to routes.
	// the controller object will be available inside the handler
	r.Handle("POST", "/user", c.UserCreate)
	r.Handle("GET", "/user/:userid/confirm", c.UserConfirm)

	r.PUT("/profile", c.ProfileUpdate)
	r.GET("/profile/:profid/history", c.ProfileHistory)

	// in this specific case httprouter can't differ urls like /:profid and /all
	// we need to handle it manually by our custom middleware forwarder
	f := r.Forward("GET", "/profile/:profid", c.ProfileGet)
	f.Forward("/profile/all", c.ProfileAll)

	f = r.GETForwarder("/profile/:profid", c.ProfileGet)
	f.Route("/profile/all", c.ProfileAll)

	fwd := server.NewForwarder("profid", c.ProfileGet)
	fwd.Route("all", c.ProfileAll)

	r.GET("/profile/:profid", fwd)

	// Fire up the server
	log.Panic(http.ListenAndServe("localhost:"+port, r))
}

// all env variable are necessary
// program stops if one is missing
func env(name string) string {
	str := os.Getenv(name)
	if env == "" {
		log.Panicf("can't find environment variable %s", name)
	}

	return str
}

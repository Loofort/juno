package controller


// Controller provides handler for each routes
// It keeps storage object
type Controller {
	storage storage.Storage
}


func (c Controller) UserCreate(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ) {

}

	r.GET("/user/:userid/confirm", c.UserConfirm)

	r.PUT("/profile", c.ProfileUpdate)
	r.GET("/profile/:profid/history", c.ProfileHistory)

	// in this specific case httprouter can't differ urls like /:profid and /all
	// we need to handle it manually by our custom middleware forwarder
	fwd := server.NewForwarder("profid", c.ProfileGet)
	fwd.Route("all", c.ProfileAll)



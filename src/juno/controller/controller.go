package controller

import (
	"golang.org/x/net/context"
	"juno/common/io"
	"juno/middle"
	"juno/model"
	"juno/model/storage"
	"net/http"
)

// Controller provides handler for each routes
// It keeps storage object
type Controller struct {
	stg storage.Storage
}

// ################ User Handlers ##################

func (c Controller) UserCreate(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	user := model.NewUser()
	if inputErr(r, user) {
		return
	}

	// Validate says which field is invalid
	if err := user.Validate(); err != nil {
		io.ErrClient(w, err)
		return
	}

	// we check dublicates on insert
	user, err := c.stg.UserInsert(ctx, user)
	if err != nil {
		if _, ok := err.(storage.ErrDublicate); ok {
			io.ErrClient(w, "The email is already registered")
			return
		}
		dbErr(w, err)
		return
	}

	// success
	resp := map[string]string{
		"message": "Please, check your mailbox for confirmation letter",
		"userid":  user.ID,
	}
	io.Output(w, resp)
}

func (c Controller) UserConfirm(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	userid, _ := middle.CtxParam("userid")

	user, err := c.stg.UserGet(ctx, userid)
	if dbErr(w, err) {
		return
	}

	if user == nil {
		io.Err(w, "User not found", http.StatusNotFound)
		return
	}

	if user.Confirm {
		io.ErrClient(w, "User is already confirmed")
		return
	}

	user.Confirm = true
	user, err = c.stg.UserUpdate(ctx, user)

	if dbErr(w, err) {
		return
	}

	// success
	resp := map[string]string{
		"message": "You are able to edit profile",
		"profid":  user.ID,
	}
	io.Output(w, resp)
}

// ################ Profile Handlers ##################

// ProfileUpdate Handler allows to modify own user profile.
func (c Controller) ProfileUpdate(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	// read profile from input
	inProfile := model.NewProfile()
	if inputErr(r, inProfile) {
		return
	}

	// Validate says which field is invalid
	if err := inProfile.Validate(); err != nil {
		io.ErrClient(w, err)
		return
	}

	// read profile from storage. W is telling that we want open profile for writing
	dbProfile, err := c.stg.ProfileGetW(ctx, inProfile.ID)
	if dbErr(w, err) {
		return
	}

	if dbProfile == nil {
		// profile doesn't exist or user hasn't permission
		// in both cases client is doing something wrong
		io.ErrClient(w, "can't update this profile")
		return
	}

	// subtract the difference
	profile, err := c.stg.ProfileUpdateHistory(ctx, dbProfile, inProfile)
	if dbErr(w, err) {
		return
	}

	io.Output(w, profile)
}

// ProfileUpdate Handler allows to view just own profile history.
func (c Controller) ProfileHistory(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	profid, _ := middle.CtxParam("profid")

	changes, err := c.stg.ProfileHistory(ctx, profid)
	if dbErr(w, err) {
		return
	}
	if changes == nil {
		// err code could be either 404 or 403, so let it be 400
		io.ErrClient(w, "can't display history of this profile")
		return
	}

	io.Output(w, changes)
}

func (c Controller) ProfileGet(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	profid, _ := middle.CtxParam("profid")

	profile, err := c.stg.ProfileGet(ctx, profid)
	if dbErr(w, err) {
		return
	}

	if profile == nil {
		// profile can be obtained by anonym, it just not found
		io.Err(w, "profile not found", http.StatusNotFound)
		return
	}

	io.Output(w, profile)
}

func (c Controller) ProfileAll(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	profiles, err := c.stg.ProfileAll(ctx, profid)
	if dbErr(w, err) {
		return
	}

	io.Output(w, profiles)
}

// ##################### helper functions ##################

func dbErr(w http.ResponseWriter, err error) bool {
	if err != nil {
		io.ErrServer(w, "Oops! database problem, try again latter")
		return true
	}
	return false
}

func inputErr(r *http.Request, obj interface{}) bool {
	if err := io.Input(r, obj); err != nil {
		io.ErrClient(w, "something wrong with your request body")
		return true
	}
	return false
}

package controller

import (
	"golang.org/x/net/context"
	"juno/common/io"
	"juno/middle"
	"juno/model"
	"juno/model/storage"
	"log"
	"net/http"
)

// error messages displayed to client
const (
	ERR_DB     = "Oops! database problem, try again latter"
	ERR_NOPROF = "profile not found"
	DB_NOUSER  = "user not found"
	ERR_REQ    = "something wrong with your request body"
)

// initialize custom loger, because we will use log.Output(calldepth, msg)
var clog = log.New(os.Stderr, "", log.LstdFlags)

// Controller provides handler for each routes
// It keeps storage object
type Controller struct {
	stg storage.Storage
}

// ################ User Handlers ##################

func (c Controller) UserCreate(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	user := &model.User{}
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
		if c.stg.IsErrDups(err) {
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

	// execute getAndModify on storage
	user, err = c.stg.UserConfirm(ctx, userid)
	if c.dbErrOrEmpty(w, err, DB_NOUSER) {
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
	inProfile := &model.Profile{}
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
	if c.dbErrOrEmpty(w, err, ERR_NOPROF) {
		return
	}

	// subtract the difference
	profile, err := c.stg.ProfileUpdateHistory(ctx, dbProfile, inProfile)
	if dbErr(w, err) {
		return
	}

	io.Output(w, profile)
}

// ProfileUpdate Handler allows to modify own user profile.
func (c Controller) ProfileUpdate(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	// read profile from input
	profile := &model.Profile{}
	if inputErr(r, profile) {
		return
	}

	// Validate says which field is invalid
	if err := profile.Validate(); err != nil {
		io.ErrClient(w, err)
		return
	}

	// update
	profile, err := c.stg.ProfileUpdate(ctx, profile)
	if c.dbErrOrEmpty(w, err, ERR_NOPROF) {
		return
	}

	io.Output(w, profile)
}

// ProfileUpdate Handler allows to view just own profile history.
func (c Controller) ProfileHistory(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	profid, _ := middle.CtxParam("profid")

	changes, err := c.stg.ProfileHistory(ctx, profid)
	if c.dbErrOrEmpty(w, err, ERR_NOPROF) {
		return
	}

	io.Output(w, changes)
}

func (c Controller) ProfileGet(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	profid, _ := middle.CtxParam("profid")

	profile, err := c.stg.ProfileGet(ctx, profid)
	if c.dbErrOrEmpty(w, err, ERR_NOPROF) {
		return
	}

	return profile
}

func (c Controller) ProfileAll(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	profiles, err := c.stg.ProfileAll(ctx)
	if dbErr(w, err) {
		return
	}

	io.Output(w, profiles)
}

// ##################### Helper Functions ##################

func (c Controller) dbErrOrEmpty(w http.ResponseWriter, err error, msg string) bool {
	if c.stg.IsErrNotFound(err) {
		io.Err(w, msg, http.StatusNotFound)
		return true
	}
	return dbErr(w, err)
}

func dbErr(w http.ResponseWriter, err error) bool {
	if err != nil {
		// send to user common err message
		io.ErrServer(w, ERR_DB)

		// log real db error
		clog.Output(2, err.Error())

		return true
	}
	return false
}

func inputErr(r *http.Request, obj interface{}) bool {
	if err := io.Input(r, obj); err != nil {
		io.ErrClient(w, ERR_REQ)
		return true
	}
	return false
}

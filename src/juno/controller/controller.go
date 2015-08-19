package controller

import (
	"golang.org/x/net/context"
	"juno/common/check"
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
	user := &model.User{}
	if check.InputErr(r, user) {
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
		check.DBErr(w, err)
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
	fields = model.Fields{"confirm": true}
	filter = model.Fields{"confirm": false}
	user, err = c.stg.UserSet(ctx, userid, fields, filter)
	if c.dbErrOrEmpty(w, err, io.ERR_NOUSER) {
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
	profile := &model.Profile{}
	if check.InputErr(r, profile) {
		return
	}

	// Validate says which field is invalid
	if err := profile.Validate(); err != nil {
		io.ErrClient(w, err)
		return
	}

	// When storage updates profile it also aupdates History, so there is two model objects have to be updated.
	// In complex program we would have to implement transaction object and would used it like :
	// tns := stg.NewTransaction(ctx)
	// tns.ProfileUpdate(profile)
	// tns.HistoryUpdate(history)
	// tns.Execute()
	profile, err := c.stg.ProfileUpdate(ctx, profile)
	if c.dbErrOrEmpty(w, err, io.ERR_NOPROF) {
		return
	}

	io.Output(w, profile)
}

// ProfileUpdate Handler allows to view just own profile history.
func (c Controller) ProfileHistory(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	profid, _ := middle.CtxParam("profid")

	changes, err := c.stg.ProfileHistory(ctx, profid)
	if c.dbErrOrEmpty(w, err, io.ERR_NOPROF) {
		return
	}

	io.Output(w, changes)
}

func (c Controller) ProfileGet(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	profid, _ := middle.CtxParam("profid")

	profile, err := c.stg.ProfileGet(ctx, profid)
	if c.dbErrOrEmpty(w, err, io.ERR_NOPROF) {
		return
	}

	return profile
}

func (c Controller) ProfileAll(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	profiles, err := c.stg.ProfileSearch(ctx, nil)
	if check.DBErr(w, err) {
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
	return check.DBErr(w, err)
}

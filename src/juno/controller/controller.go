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

func New(stg storage.Storage) Controller {
	return Controller{stg}
}

// ################ User Handlers ##################

func (c Controller) UserCreate(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	user := &model.User{}
	if check.InputErr(w, r, user) {
		return
	}

	// Validate says which field is invalid
	if msg := user.Validate(); msg != "" {
		io.ErrClient(w, msg)
		return
	}

	// we check dublicates on insert
	user, err := c.stg.UserInsert(ctx, user)
	if err != nil {
		if c.stg.IsErrDup(err) {
			io.ErrClient(w, "The email is already registered")
			return
		}
		check.DBErr(w, err)
		return
	}

	// success
	resp := map[string]string{
		"message": "Please, check your mailbox for confirmation letter",
		"id":      user.ID,
	}
	io.Output(w, resp)
}

func (c Controller) UserConfirm(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	userid, _ := middle.CtxParam(ctx, "userid")

	// execute getAndModify on storage
	fields := model.Fields{"confirm": true}
	filter := model.Fields{"confirm": false}
	user, err := c.stg.UserSet(ctx, userid, fields, filter)
	if c.dbErrOrEmpty(w, err, io.ERR_NOUSER) {
		return
	}

	// success
	resp := map[string]string{
		"message": "You are able to edit profile",
		"id":      user.ID,
	}
	io.Output(w, resp)
}

// ################ Profile Handlers ##################

// ProfileUpdate Handler allows to modify own user profile.
func (c Controller) ProfileUpdate(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	// read profile from input
	profile := &model.Profile{}
	if check.InputErr(w, r, profile) {
		return
	}

	// Validate says which field is invalid
	if msg := profile.Validate(); msg != "" {
		io.ErrClient(w, msg)
		return
	}

	// When storage updates profile it also aupdates History, so there is two model objects have to be updated.
	// In complex program we would have to implement transaction object and would used it like :
	// txn := stg.NewTransaction(ctx)
	// txn.ProfileUpdate(profile)
	// txn.HistoryUpdate(history)
	// txn.Execute()
	profile, err := c.stg.ProfileUpdate(ctx, profile)
	if c.dbErrOrEmpty(w, err, io.ERR_NOPROF) {
		return
	}

	io.Output(w, profile)
}

// ProfileUpdate Handler allows to view just own profile history.
func (c Controller) ProfileHistory(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	profid, _ := middle.CtxParam(ctx, "profid")

	changes, err := c.stg.HistoryGet(ctx, profid)
	if c.dbErrOrEmpty(w, err, io.ERR_NOPROF) {
		return
	}

	io.Output(w, changes)
}

func (c Controller) ProfileGet(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	profid, _ := middle.CtxParam(ctx, "profid")

	profile, err := c.stg.ProfileGet(ctx, profid)
	if c.dbErrOrEmpty(w, err, io.ERR_NOPROF) {
		return
	}

	io.Output(w, profile)
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

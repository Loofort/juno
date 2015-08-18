package storage

import (
	"golang.org/x/net/context"
	"gopkg.in/mgo.v2/bson"
	"juno/model"
)

// User represents mongo specific fields for model User
type UserDB struct {
	ID         bson.ObjectId `bson:"_id,omitempty"`
	model.User `bson:"inline"`
}

func (db *UserDB) Model() *model.User {
	db.User.ID = db.ID.Hex()
	return &db.User
}

// Profile represents mongo specific fields for model Profile
type ProfileDB struct {
	ID      bson.ObjectId `bson:"_id,omitempty"`
	Profile model.Profile `bson:"profile"`
}

func (db *ProfileDB) Model() *model.Profile {
	db.Profile.ID = db.ID.Hex()
	return &db.Profile
}

// ChangesDB is the mongo specific wrapper for model []Change
type ChangesDB struct {
	ProfileDB `bson:"changes"`
	Changes   []*model.Change `bson:"changes"`
}

func (db *ChangesDB) Model() []*model.Change {
	return db.Changes
}

// Storage represent DB layer, it is aware of model, but model doesn't aware of storage
// For now only mongoDB is available
type Storage interface {
	Reserve(ctx context.Context) (context.Context, ReleaseFunc)
	Close()
	IsErrNotFound(error) bool
	IsErrDup(error) bool

	// ############## User Section ########################
	UserByCreds(ctx context.Context, email, pass string) (*model.User, error)
	UserInsert(ctx context.Context, user *model.User) (*model.User, error)
	UserGet(ctx context.Context, userid string) (*model.User, error)
	UserUpdate(ctx context.Context, user *model.User) (*model.User, error)

	// ############## Profile Section ###################
	ProfileAll(ctx context.Context) ([]*model.Profile, error)
	ProfileGet(ctx context.Context, profid string) (*model.Profile, error)
	ProfileGetW(ctx context.Context, profid string) (*model.Profile, error)
	ProfileUpdateHistory(ctx context.Context, prev, curr *model.Profile) (*model.Profile, error)
	ProfileHistory(ctx context.Context, profid string) (*[]model.Change, error)
}

// type of function that release db resourses
type ReleaseFunc func()

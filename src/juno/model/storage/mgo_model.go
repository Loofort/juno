package storage

import (
	"golang.org/x/net/context"
	"gopkg.in/mgo.v2/bson"
	"juno/model"
)

// ModelDB represent structure for all BL objects
type ModelDB struct {
	ID         bson.ObjectId `bson:"_id"`
	model.User `bson:"inline"`
	Profile    model.Profile
	Changes    []*model.Change
}

// User represents mongo specific fields for model User
type UserDB struct {
	ID         bson.ObjectId `bson:"_id"`
	model.User `bson:"inline"`
}

func (db *UserDB) Model() *model.User {
	db.User.ID = db.ID.Hex()
	return &db.User
}

// Profile represents mongo specific fields for model Profile
type ProfileDB struct {
	ID      bson.ObjectId `bson:"_id"`
	Profile model.Profile
}

func (db *ProfileDB) Model() *model.Profile {
	db.Profile.ID = db.ID.Hex()
	return &db.Profile
}

// ChangesDB is the mongo specific wrapper for model []Change
type ChangesDB struct {
	ID      bson.ObjectId `bson:"_id"`
	Changes []*model.Change
}

func (db *ChangesDB) Model() []*model.Change {
	return db.Profile.Changes
}

// storage represents CRUD-like operation for each object
// it is aware of model, but model doesn't aware of storage
// For now only mongoDB is available
type Storage interface {
	Reserve(ctx context.Context) (context.Context, ReleaseFunc)
	Close()
	IsErrNotFound(error) bool
	IsErrDup(error) bool

	// ############## User Section ########################
	UserSearch(ctx context.Context, filter model.Fields) (*model.User, error)
	UserInsert(ctx context.Context, user *model.User) (*model.User, error)
	UserGet(ctx context.Context, userid string) (*model.User, error)
	UserSet(ctx context.Context, fields, filter model.Fields) (*model.User, error)

	// ############## Profile Section ###################
	ProfileSearch(ctx context.Context, filter model.M) ([]*model.Profile, error)
	ProfileGet(ctx context.Context, profid string) (*model.Profile, error)
	ProfileUpdate(ctx context.Context, profile *model.Profile) (*model.Profile, error)

	// ############## History Section ###################
	HistoryGet(ctx context.Context, histid string) (*model.History, error)
}

// type of function that release db resourses
type ReleaseFunc func()

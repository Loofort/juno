package storage

import (
	"fmt"
	"golang.org/x/net/context"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"juno/middle"
	"log"
)

const MGO_COLLECTION = "people"

type mongoStg struct {
	// we will copy the base session each time we need concurrent mgo call
	base *mgo.Session
}

// MgoMustConnect create mongo connection, and returns storage.
// It's intended to be called on startup, as storage constructor.
// It panics on error
func MgoMustConnect(mhost, database, name, pass string) Storage {
	// connect to mongo
	// todo: create a separate mongo user for particular database
	info := &mgo.DialInfo{
		Addrs:    []string{mhost},
		Timeout:  60 * time.Second,
		Database: database,
		Username: name,
		Password: pass,
	}

	sess, err := mgo.DialWithInfo(info)
	if err != nil {
		panic(err)
	}

	// Optional. Switch the session to a monotonic behavior.
	sess.SetMode(mgo.Monotonic, true)
	c := sess.DB().C(MGO_COLLECTION)

	// create indexes if don't exist
	indexes := []Index{
		// to control email duplicates,
		Index{
			Key:        []string{"email"},
			Unique:     true,
			DropDups:   false,
			Background: true,
			Sparse:     true,
		},
		// for fast auth access
		Index{
			Key:        []string{"email", "password"},
			DropDups:   false,
			Background: true,
			Sparse:     true,
		},
		// to get confirmed user
		Index{
			Key:        []string{"_id", "confirm"},
			Unique:     true,
			Background: true,
			Sparse:     true,
		},
	}

	for _, index = range indexes {
		if err := c.EnsureIndex(index); err != nil {
			panic(err)
		}
	}

	return &mongoStg{sess}
}

// Close is shutdown the connection,
// It's intended to be call as destructor
func (s mongoStg) Close() {
	s.sess.Close()
}

func (s mongoStg) IsErrNotFound(err error) bool {
	return err == mgo.ErrNotFound
}
func (s mongoStg) IsErrDup(err error) bool {
	return mgo.IsDup(err)
}

// ################### functions for context #################
type ctxKey int

var colKey ctxKey = 0

// Reserve spawns db session copy
// and puts in context collection that could be queried concurrently.
// It returns modified context and release function that puts db session back to the pool
func (s mongoStg) Reserve(ctx context.Context) (context.Context, ReleaseFunc) {
	sess = s.base.Copy()
	c := sess.DB().C(MGO_COLLECTION)

	// in complex application we would preserve session, not collection
	ctx = context.WithValue(ctx, colKey, c)
	release := func() {
		sess.Close()
	}

	return ctx, release
}

// col return collection preserved in context
func col(ctx context.Context) *mgo.Collection {
	c, ok := ctx.Value(colKey).(*mgo.Collection)
	if !ok {
		log.Println("no collection in context") // todo: write call stack
		c = s.base.DB().C(MGO_COLLECTION)
	}
	return c
}

// ###################### User CRUD Section #########################

// todo: implement permission check to all methods:
//  get current user from context (may be anonym).
//  get user roles array and find objects matched with roles.
//  each object has two role arrays "writes" and "reads", depending on method we need to go through one of them
//  finite role array can be encoded as bit string.
//  additionally we need check ownership by $or clause

// UserByCreds Looks for user by email and password
func (s mongoStg) UserSearch(ctx context.Context, filter model.Fields) (*model.User, error) {
	udb := &UserDB{}
	err := col(ctx).Find(bson.M(filter)).One(udb)
	return udb.Model(), err
}

// UserInsert creates new user. it overrides ID if any
func (s mongoStg) UserInsert(ctx context.Context, userm *model.User) (*model.User, error) {
	user := &UserDB{
		User: *userm,
		ID:   bson.NewObjectId(),
	}

	err := col(ctx).Insert(user)
	return user.Model(), err
}

func (s mongoStg) UserGet(ctx context.Context, userid string) (*model.User, error) {
	user := &UserDB{}
	err := getByID(ctx, userid, user, nil)

	return user.Model(), err
}

// UserSet gets user applying optional filter and modifies the object
// it will rise ErrNotFound if user is already confirmed
func (s mongoStg) UserSet(ctx context.Context, userid string, fields, filter model.Fields) (*model.User, error) {
	id, err := toObjectId(userid)
	if err != nil {
		return nil, err
	}

	if filter == nil {
		filter = model.Fields{}
	}
	filter["_id"] = id

	c := col(ctx)
	err = c.Update(bson.M(filter), bson.M{"$set": bson.M(fields)})
	if err != nil {
		return nil, err
	}

	user := &UserDB{}
	err = c.FindId(id).One(user)
	return user, err
}

// ########################## Profile CRUD Section ##############################

// ProfileSearch obtains profiles of confirmed users. It limits result (to 1k) for security reasons.
func (s mongoStg) ProfileSearch(ctx context.Context, fields model.Fields) ([]*model.Profile, error) {

	pdbs := []*ProfileDB{}
	query := col(ctx).Find(confirm(filter)).Limit(1000)
	if err := query.All(&pdbs); err != nil {
		return nil, err
	}

	// convert db profiles to model profiles
	profiles = make([]*model.Profile, 0, len(pdbs))
	for _, p = range pdbs {
		profiles = append(profiles, p.Model())
	}

	return profiles, err
}

func (s mongoStg) ProfileGet(ctx context.Context, profid string) (*model.Profile, error) {
	item := &ProfileDB{}
	err := getByID(ctx, profid, item, confirm(nil))

	return item.Model(), err
}

// ProfileUpdate updates profile and saves history changes
func (s mongoStg) ProfileUpdate(ctx context.Context, profile *model.Profile) ([]*model.Profile, error) {

	// check permissions.
	// todo: remove this crutch if common permission workflow is implemented
	if err := requestAccess(ctx, profile.ID); err != nil {
		return nil, err
	}

	prev, err := s.ProfileGet(ctx, profile.ID)
	if err != nil {
		return nil, err
	}

	// Substract profile changes
	change := prev.Substract(profile)

	update := bson.M{
		"$set": bson.M{
			"profile": profile,
		},
		"$push": bson.M{
			"changes": change,
		},
	}

	filter := confirm(nil)
	filter["_id"] = toObjectId(profile.ID)

	err = c.Update(filter, update)

	return profile, err
}

// ################ History CRUD section ####################

// HistoryGet requests changes on behalf of context user.
func (s mongoStg) HistoryGet(ctx context.Context, profid string) ([]*model.Change, error) {
	// check permissions
	if err := requestAccess(ctx, profid); err != nil {
		return nil, err
	}

	changes := []ChangesDB{}
	err := getByID(ctx, profid, &changes, confirm(nil))
	return changes.Model(), err
}

// ############### helper functions #################

// confirm adds to filter confirm clause
func confirm(filter model.Fields) bson.M {
	if filter == nil {
		filter = model.Fields{}
	}
	filter["confirm"] = true
	return bson.M(filter)
}

// toObjectId checks string first because ObjectIdHex(id) panics on incorrect input
func toObjectId(id string) (bson.ObjectId, error) {
	if !bson.IsObjectIdHex(id) {
		return bson.ObjectId{}, fmt.Errorf("Can't convert to ObjectId %s", id)
	}

	return bson.ObjectIdHex(id), nil
}

// fetch mongo object by string id
func getByID(ctx context.Context, id string, obj interface{}, filter bson.M) error {
	oid, err := toObjectId(id)
	if err != nil {
		return nil, err
	}

	if filter == nil {
		filter = bson.M{}
	}
	filter["_id"] = oid

	return col(ctx).Find(filter).One(obj)
}

// requestAccess checks if context user is entitled to access the object
func requestAccess(ctx context.Context, id string) error {
	user := middle.CtxUser(ctx)
	if user.ID != id {
		return nil, mgo.ErrNotFound
	}
}

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
	}

	for _, index = range indexes {
		if err := c.EnsureIndex(uniqEmail); err != nil {
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

// ###################### User Section #########################

// todo: implement permission check to all methods:
//  get current user from context (may be anonym).
//  get user roles array and find objects matched with roles.
//  each object has two role arrays "writes" and "reads", depending on method we need to go through one of them
//  finite role array can be encoded as bit string.
//  additionally we need check ownership by $or clause

// UserByCreds Looks for user by email and password
func (s mongoStg) UserByCreds(ctx context.Context, email, password string) (*model.User, error) {
	udb := &UserDB{}
	filter := bson.M{"email": email, "password": password}

	err := col(ctx).Find(filter).One(udb)
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
	err := getByID(ctx, userid, user)

	return user.Model(), err
}

// UserConfirm find and modify user, it will rise ErrNotFound if user is already confirmed
func (s mongoStg) UserConfirm(ctx context.Context, userid string) (*model.User, error) {
	id, err := toObjectId(userid)
	if err != nil {
		return nil, err
	}

	c := col(ctx)
	err = c.Update(bson.M{"_id": id, "confirm": false}, bson.M{"confirm": true})
	if err != nil {
		return nil, err
	}

	user := &UserDB{}
	err = c.FindId(id).One(user)
	return user, err
}

// ########################## Profile Section ##############################

// ProfileAll obtains profiles of confirmed users. It limits result (to 1k) for security reasons.
func (s mongoStg) ProfileAll(ctx context.Context) ([]*model.Profile, error) {

	query := col(ctx).Find(bson.M{"confirm": true})
	query = query.Select(bson.M{"profile": 1}).Limit(1000)

	pdbs := []*ProfileDB{}
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
	err := getByID(ctx, profid, item)

	return item.Model(), err
}

// ProfileGetW looks for the profile on behalf of context user.
// If profile is not found, It could mean user doesn't have write permisson for the profile
func (s mongoStg) ProfileGetW(ctx context.Context, profid string) (*model.Profile, error) {

	// check permissions.
	// todo: remove this crutch if common permission workflow was implemented
	if err := requestAccess(ctx, profid); err != nil {
		return nil, err
	}

	return s.ProfileGet(ctx, profid)
}

// ProfileHistory requests changes on behalf of context user.
func (s mongoStg) ProfileHistory(ctx context.Context, profid string) (*model.Profile, error) {
	// check permissions
	if err := requestAccess(ctx, profid); err != nil {
		return nil, err
	}

	changes := []ChangesDB{}
	err := getByID(ctx, profid, &changes)
	return changes.Model(), err
}

// ProfileUpdateHistory substracts current and next profiles and save profile with changes
func (s mongoStg) ProfileUpdateHistory(ctx context.Context, curr, next *model.Profile) ([]*model.Change, error) {
}

// ############### helper functions #################

// toObjectId checks string first because ObjectIdHex(id) panics on incorrect input
func toObjectId(id string) (bson.ObjectId, error) {
	if !bson.IsObjectIdHex(id) {
		return bson.ObjectId{}, fmt.Errorf("Can't convert to ObjectId %s", id)
	}

	return bson.ObjectIdHex(id), nil
}

// fetch mongo object by string id
func getByID(ctx context.Context, id string, obj interface{}) error {
	oid, err := toObjectId(id)
	if err != nil {
		return nil, err
	}

	return col(ctx).FindId(oid).One(obj)
}

// requestAccess checks if context user is entitled to access the object
func requestAccess(ctx context.Context, id string) error {
	user := middle.CtxUser(ctx)
	if user.ID != id {
		return nil, mgo.ErrNotFound
	}
}

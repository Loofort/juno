package model

import (
	"golang.org/x/net/context"
	"log"
	"time"
)

const ANONYM_ID = "anonym_id"

// the anonym has prefilled privileges
var anonym = User{
	ID: ANONYM_ID,
}

// Anonym returns a copy of special anonym user
func Anonym() *User {
	copyUser := anonym
	return &copyUser
}

// Fields represents an arbitrary set of object fields,
// used mainly in storage calls
type Fields map[string]interface{}

// User represents user object.
// it contains authentication and identification data (like login, password),
// it also might contain in future auth tokens, permissions (roles)
type User struct {
	// string represent of ID that uses by http requests
	ID string `bson:"-"`
	// email is used for confirmation letter (not implemented) and as login
	Email string
	// Password currently is stored as plain text, even without md5.
	// in general it should be preserved using some strong crypto algo (e.g. HMAC)
	Password string

	Confirm bool
}

func (u *User) Validate() string {
	// todo:
	return ""
}

// Profile represnts editable user profile data
type Profile struct {
	ID        string `bson:"-"`
	FirstName string
	LastName  string
	Address   string
	Phone     string
	Age       int
}

func (p *Profile) Validate() string {
	// todo:
	return ""
}

// Substract calculates difference between current and next profile version
// And returns change object
func (p *Profile) Substract(next *Profile) Change {
	change := Change{Time: time.Now()}

	// calulate each field manualy
	// todo: for large object use reflection
	fields := make(map[string]ChangedField, 5)
	if p.FirstName != next.FirstName {
		fields["FirstName"] = ChangedField{p.FirstName, next.FirstName}
	}
	if p.LastName != next.LastName {
		fields["LastName"] = ChangedField{p.LastName, next.LastName}
	}
	if p.Address != next.Address {
		fields["Address"] = ChangedField{p.Address, next.Address}
	}
	if p.Phone != next.Phone {
		fields["Phone"] = ChangedField{p.Phone, next.Phone}
	}
	if p.Age != next.Age {
		fields["Age"] = ChangedField{p.Age, next.Age}
	}

	change.Fields = fields
	return change
}

// Change represents one history change of profile.
// It contains previous and current value for each changed prfofile field.
type Change struct {
	Time   time.Time
	Fields map[string]ChangedField
}

// ChangedField represents changed field preserved in history.
// it contains previous and current value
type ChangedField struct {
	Previous interface{}
	Current  interface{}
}

// To avoid key collisions in context we defines an unexported type key
type ctxKey int

var userKey ctxKey = 0

// setCtxUser adds user object to conext
func SetCtxUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// CtxUser returns User from context
func CtxUser(ctx context.Context) *User {
	user, ok := ctx.Value(userKey).(*User)
	if !ok {
		log.Println("no user in context") // todo: write call stack
		user = Anonym()
	}
	return user
}

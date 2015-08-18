package model

const ANONYM_ID = "anonym_id"

// the anonym has prefilled privileges
var anonym = User{
	ID: ANONYM_ID,
}

// Anonym returns a copy of special anonym user
func Anonym() *model.User {
	copyUser := user
	return &copyUser
}

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

func (u User) Validate() string {
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

func (p Profile) Validate() string {
	// todo:
	return ""
}

// Change represents one history change of profile.
// It contains previous and current value for each changed prfofile field.
type Change struct {
	Timestamp int
	Fields    map[string]ChangedField
}

// ChangedField represents changed field preserved in history.
// it contains previous and current value
type ChangedField struct {
	Previous interface{}
	Current  interface{}
}

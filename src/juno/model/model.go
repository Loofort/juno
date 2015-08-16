package model

// CheckPerms check whether the user has one of the allowed roles
func CheckPerms(user User, owner string, allowed []int) bool {
	return user.ID == owner
	// todo: check roles
}


if !user.HasPerms(dbProfile.ID, dbProfile.Writes) {
		// user  doesn't exist, user is school hacker
		common.ErrClient(w, ERR_INPUT)
		return
	}

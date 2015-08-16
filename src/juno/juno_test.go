package main

import (
	"github.com/bndr/gopencils"
	"os"
	"testing"
	"time"
)

var apiurl string = "http://localhost" + os.Getenv("JUNO_PORT") + "/v1/"

// acceptance test for juno server. Server should be up and running
func LiveCircleTest(t *testing.T) {

	now := time.Now().String()
	email, pass := "test"+now+"@mail.com", "pass"+now
	auth := &gopencils.BasicAuth{email, pass}

	//############# register and confirm new user ################
	userid, err := createUser(email, pass)
	if err != nil {
		t.Fatal(err)
	}

	_, err = createUser(email, pass)
	if err == nil {
		t.Fatal("second registration should be considered as an error")
	}

	_, err = getProfile(userid)
	if err == nil {
		t.Fatal("unconfirmed user shouldn't have profile")
	}

	profid, err := confirm(userid)
	if err != nil {
		t.Fatal(err)
	}

	_, err = confirm(userid)
	if err == nil {
		t.Fatal("second confirmation should be considered as an error")
	}

	//############### fill up and update profile ##############
	profile, err := getProfile(profid)
	if err != nil {
		t.Fatal(err)
	}

	profile.FirstName = "John"
	profile.LastName = "Smith"
	profile.Address = "100 E. 17th Street, New York"
	profile.Phone = "+1-212-674-4300"
	profile, err = updateProfile(auth, profile)
	if err != nil {
		t.Fatal(err)
	}

	profile.FirstName = "Will"
	profile, err = updateProfile(auth, profile)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("final profile for %s: %v", email, profile)
	if profile.FirstName != "Will" {
		t.Fatal("profile hasn't been updated")
	}

	changes, err := getHistory(auth, profid)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("profile changes for %s: %v", email, changes)
	if len(changes) != 2 {
		t.Fatalf("history changes: expected 2, but get %d", len(changes))
	}

	// ########## check anonymoues access ############
	profiles, err := allProfiles()
	if err != nil {
		t.Fatal(err)
	}
	if len(profiles) < 1 {
		t.Fatal("can't obtain profiles")
	}

	profid = profiles[0].ID
	profile, err = getProfile(profid)
	if err != nil {
		t.Fatal(err)
	}

	profile.FirstName = "Anon"
	profile, err = updateProfile(nil, profile)
	if err == nil {
		t.Fatalf("anon update profile: expected error, but get profile %v", profile)
	}

	changes, err = getHistory(nil, profid)
	if err != nil {
		t.Fatalf("anon view history: expected error, but get changes %v", changes)
	}
}

func ArbitraryAccessTest(t *testing.T) {
	userid := time.Now().String()
	_, err := confirm(userid)
	if err == nil {
		t.Fatal("Arbitrary confirm should return error")
	}

	_, err = getProfile(userid)
	if err == nil {
		t.Fatal("Arbitrary profile access should return error")
	}
}

func CrossAccessTest(t *testing.T) {
	now := time.Now().String()
	email1, pass1 := "test1"+now+"@mail.com", "pass1"+now
	email2, pass2 := "test2"+now+"@mail.com", "pass2"+now

	auth1, profile1 := register(t, email1, pass1)
	auth2, profile2 := register(t, email2, pass2)

	profile1.FirstName = "hacked"
	profile1, err = updateProfile(auth2, profile1)
	if err == nil {
		t.Fatalf("cross update profile: expected error, but get profile %v", profile1)
	}

	changes, err = getHistory(auth1, profile2.ID)
	if err != nil {
		t.Fatalf("cross view history: expected error, but get changes %v", changes)
	}
}

// ############################ Help Functions ####################################

// register creates new fake user
func register(t *testing.T, email, pass string) (*gopencils.BasicAuth, *model.Profile) {
	auth := &gopencils.BasicAuth{email, pass}

	userid, err := createUser(email, pass)
	if err != nil {
		t.Fatal(err)
	}

	profid, err := confirm(userid)
	if err != nil {
		t.Fatal(err)
	}
	profile := &model.Profile{
		ID:        profid,
		FirstName: "user name",
	}

	profile, err = updateProfile(auth, profile)
	if err != nil {
		t.Fatal(err)
	}

	return auth, profile
}

// returns user id
func createUser(email, pass) (string, error) {
	api := gopencils.Api(apiurl)
	user := &model.User{
		Email:    email,
		Password: pass,
	}

	outmap := map[string]interface{}{}
	res, err := api.Res("user", outmap).Post(user)

	if err = checkErr(res, err); err != nil {
		return "", err
	}

	return outmap["id"].(string), nil
}

// returns profile id
func confirm(userid string) (string, error) {
	api := gopencils.Api(apiurl)

	outmap := map[string]interface{}{}
	res, err := api.Res("user", outmap).Id(userid).Get()

	if err = checkErr(res, err); err != nil {
		return "", err
	}

	return outmap["id"].(string), nil
}

func allProfiles() ([]*model.Profile, error) {
	api := gopencils.Api(apiurl)

	profiles := []*model.Profile{}
	res, err := api.Res("profile", &profiles).Get()

	if err = checkErr(res, err); err != nil {
		return nil, err
	}

	return profiles, nil
}

func getProfile(profid string) (*model.Profile, err) {
	api := gopencils.Api(apiurl)

	profile := &model.Profile{}
	res, err := api.Res("profile", profile).Id(profid).Get()

	if err = checkErr(res, err); err != nil {
		return nil, err
	}

	return profile, nil
}

func updateProfile(auth *gopencils.BasicAuth, profile *model.Profile) (*model.Profile, err) {
	api := gopencils.Api(apiurl, auth)

	updatedProfile := &model.Profile{}
	res, err := api.Res("profile", updatedProfile).Put(profile)

	if err = checkErr(res, err); err != nil {
		return nil, err
	}

	return profile, nil
}

func getHistory(auth *gopencils.BasicAuth, profid string) ([]*model.Change, error) {
	api := gopencils.Api(apiurl, auth)

	changes := []*model.Change{}
	res, err := api.Res("profile", &changes).Id(profid).Res("history", changes).Get()

	if err = checkErr(res, err); err != nil {
		return nil, err
	}

	return changes, nil
}

type ErrorResp struct {
	Code    int
	Message string
}

func checkErr(res *gopencils.Resource, err error) error {
	if err != nil {
		return err
	}

	eresp := &ErrorResp{}
	err = json.NewDecoder(res.Row.Body).Decode(eresp)
	if err != nil {
		return err
	}

}

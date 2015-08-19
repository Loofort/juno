package main

import (
	"fmt"
	"github.com/bndr/gopencils"
	"juno/common/io"
	"juno/model"
	"log"
	"os"
	"strconv"
	"testing"
	"time"
)

var apiurl string = fmt.Sprintf("http://localhost:%s/%s", os.Getenv("JUNO_PORT"), VER)

// acceptance test for juno server. Server should be up and running
func TestJunoLiveCircle(t *testing.T) {

	sufix := rand()
	email, pass := "user"+sufix+"@mail.com", "pass"+sufix
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
	profile.Address = "100 E. 17th Street, New&nbsp;<York>"
	profile.Phone = "+1-212-674-4300"
	profile.Age = 30
	profile, err = updateProfile(auth, profile)
	if err != nil {
		t.Fatal(err)
	}

	profile.FirstName = "Will"
	profile, err = updateProfile(auth, profile)
	if err != nil {
		t.Fatal(err)
	}

	if profile.FirstName != "Will" {
		t.Fatal("profile hasn't been updated")
	}

	changes, err := getHistory(auth, profid)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("profile changes: %v", changes)
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

	profile = profiles[0]
	profile.FirstName = "Anon"
	profile, err = updateProfile(nil, profile)
	if err == nil {
		t.Fatalf("anon update profile: expected error, but get profile %v", profile)
	}

	profile, err = getProfile(profiles[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	changes, err = getHistory(nil, profid)
	if err == nil {
		t.Fatalf("anon view history: expected error, but get changes %v", changes)
	}
}

func TestJunoArbitraryAccess(t *testing.T) {
	userid := rand()
	_, err := confirm(userid)
	if err == nil {
		t.Fatal("Arbitrary confirm should return error")
	}

	_, err = getProfile(userid)
	if err == nil {
		t.Fatal("Arbitrary profile access should return error")
	}
}

func TestJunoCrossAccess(t *testing.T) {
	sufix := rand()
	email1, pass1 := "test1"+sufix+"@mail.com", "pass1"+sufix
	email2, pass2 := "test2"+sufix+"@mail.com", "pass2"+sufix

	auth1, profile1 := register(t, email1, pass1)
	auth2, profile2 := register(t, email2, pass2)

	profile1.FirstName = "hacked"

	profile, err := updateProfile(auth2, profile1)
	if err == nil {
		t.Fatalf("cross update profile: expected error, but get profile %v", profile)
	}

	changes, err := getHistory(auth1, profile2.ID)
	if err == nil {
		t.Fatalf("cross view history: expected error, but get changes %v", changes)
	}
}

// ############################ Help Functions ####################################

// rand returns arbitrary string based on time
func rand() string {
	return strconv.FormatInt(time.Now().UnixNano(), 16)
}

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
func createUser(email, pass string) (string, error) {
	api := gopencils.Api(apiurl)
	user := &model.User{
		Email:    email,
		Password: pass,
	}

	outmap := map[string]interface{}{}
	res, err := api.Res("user", &outmap).Post(user)

	if err = checkErr(res, err); err != nil {
		return "", err
	}

	id, ok := outmap["id"].(string)
	if !ok {
		return "", fmt.Errorf("can't get user id, resp: %#v", outmap)
	}

	return id, nil
}

// returns profile id
func confirm(userid string) (string, error) {
	api := gopencils.Api(apiurl)

	outmap := map[string]interface{}{}
	res, err := api.Res("user").Id(userid).Res("confirm", &outmap).Get()

	if err = checkErr(res, err); err != nil {
		return "", err
	}

	id, ok := outmap["id"].(string)
	if !ok {
		return "", fmt.Errorf("can't get user id, resp: %#v", outmap)
	}

	return id, nil
}

func allProfiles() ([]*model.Profile, error) {
	api := gopencils.Api(apiurl)

	profiles := []*model.Profile{}
	res, err := api.Res("profile").Res("all", &profiles).Get()

	if err = checkErr(res, err); err != nil {
		return nil, err
	}

	return profiles, nil
}

func getProfile(profid string) (*model.Profile, error) {
	api := gopencils.Api(apiurl)

	profile := &model.Profile{}
	res, err := api.Res("profile", profile).Id(profid).Get()

	if err = checkErr(res, err); err != nil {
		return nil, err
	}

	return profile, nil
}

func updateProfile(auth *gopencils.BasicAuth, profile *model.Profile) (*model.Profile, error) {
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
	res, err := api.Res("profile").Id(profid).Res("history", &changes).Get()

	if err = checkErr(res, err); err != nil {
		return nil, err
	}

	return changes, nil
}

func checkErr(res *gopencils.Resource, err error) error {
	if err != nil {
		log.Printf("err in checkErr %v", err)
		return err
	}

	if res.Raw.StatusCode >= 400 {
		msg := res.Raw.Header.Get(io.JUNO_ERR_HEADER)
		return fmt.Errorf("err %d: %s", res.Raw.StatusCode, msg)
	}

	return nil
}

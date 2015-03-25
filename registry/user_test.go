package registry

import (
	"bytes"
	"testing"

	"github.com/acasajus/menac/db"
)

func init() {
	if err := getDB().RegisterIndexes(&User{}); err != nil && !db.IsErrIndexExists(err) {
		panic(err)
	}
}

func TestUserCreate(t *testing.T) {
	o := getDummyOrg()
	u := o.NewUser()
	u.Handle = "ASDSAD"
	u.Email = []string{"asd"}
	u.Name = "ASD"
	u.SetPassword("asdas")
	if err := u.Create(); err != nil {
		t.Fatal(err)
	}
	if err := u.Create(); err == nil || !db.IsErrDuplicateKey(err) {
		t.Fatalf("Unexpected error %s", err)
	}
	u2, err := o.GetUser(u.Handle)
	if err != nil {
		t.Fatal(err)
	}
	if u.Handle != u2.Handle || !bytes.Equal(u.Password, u2.Password) || u.Organization != u2.Organization {
		t.Errorf("Users differ %#v %#v", u, u2)
	}
	found := false
	users, err := o.Users()
	if err != nil {
		t.Fatal(err)
	}
	for _, iu := range users {
		if iu.Handle == u.Handle {
			found = true
			break
		}
	}
	if !found {
		t.Error("User could not be found in the list of org users")
	}
}

func TestUserPassword(t *testing.T) {
	u := &User{}
	pass := "DUMMYPASS"
	u.SetPassword(pass)
	if !u.CheckPassword(pass) {
		t.Fatal("Passwords do not match")
	}
	if u.CheckPassword(pass + "ASD") {
		t.Fatal("Different passwords match")
	}
}

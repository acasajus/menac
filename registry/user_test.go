package registry

import "testing"

func init() {
	if err := getDB().RegisterIndexes(&User{}); err != nil {
		panic(err)
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

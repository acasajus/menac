package registry

import (
	"testing"
	"time"

	"github.com/acasajus/menac/db"
)

var gdb db.DB

func getDB() db.DB {
	if gdb == nil {
		d, err := db.NewTestDB("127.0.0.1", 3000)
		if err != nil {
			panic(err)
		}
		gdb = d
	}
	return gdb
}

func getDummyOrg() *Organization {
	o := NewOrg(getDB())
	o.Handle = "dummyorg:" + time.Now().Format(time.RFC3339Nano)
	if err := o.Create(); err != nil {
		panic(err)
	}
	return o
}

func TestOrgCreate(t *testing.T) {
	o := NewOrg(getDB())
	if err := o.Validate(); err == nil {
		t.Error("Could validate empty org")
	}
	if err := o.Create(); err == nil {
		t.Error("Could create empty org")
	}
	o.Handle = "testorg:" + time.Now().Format(time.RFC3339Nano)
	if err := o.Create(); err != nil {
		t.Error(err)
	}
	if err := o.Create(); !db.IsErrDuplicateKey(err) {
		t.Error(err)
	}
}

func TestOrgCreateGroup(t *testing.T) {
	o := getDummyOrg()
	if _, err := o.CreateGroup("testgroup"); err != nil {
		t.Fatalf("Could not create group: %s", err)
	}
	if _, err := o.CreateGroup("testgroup"); err.Error() != "Group already exists" {
		t.Errorf("Could not create group: %s", err)
	}
	if _, err := o.CreateGroup("testgroup2"); err != nil {
		t.Errorf("Could not create group: %s", err)
	}
	g, err := o.GetGroup("testgroup")
	if err != nil {
		t.Fatalf("Could not retrieve group: %s", err)
	}
	if g.Name != "testgroup" || g.Organization != o.Handle {
		t.Error("Name or org differ")
	}
	if _, err = o.GetGroup("NONEXISTAT"); err != db.ERR_NO_EXIST {
		t.Errorf("Unexpected error: %s", err)
	}
}

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

func TestCreateOrg(t *testing.T) {
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

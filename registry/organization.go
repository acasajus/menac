package registry

import (
	"errors"
	"fmt"

	"github.com/acasajus/menac/db"
)

type Organization struct {
	db.Record

	Handle string `db:"pk"`
}

func NewOrg(db db.DB) *Organization {
	return db.LinkRecordToDB(&Organization{}).(*Organization)
}

func (o *Organization) Create() error {
	return o.GetDB().CreateNewRecord(o)
}

func (o *Organization) Validate() error {
	if len(o.Handle) == 0 {
		return errors.New("Empty organization handle")
	}
	return nil
}

func (o *Organization) NewUser() *User {
	return o.GetDB().LinkRecordToDB(&User{Organization: o.Handle}).(*User)
}

func (o *Organization) Users() ([]*User, error) {
	users := make([]*User, 0)
	u := &User{}
	for sr := range o.GetDB().Search(u, "Organization", o.Handle) {
		if sr.Error != nil {
			return nil, sr.Error
		}
		u, ok := sr.Record.(*User)
		if !ok {
			panic(fmt.Sprintf("Unexpected struct type came out of the pipe %#v", sr.Record))
		}
		users = append(users, u)
	}
	return users, nil
}

package registry

import (
	"errors"
	"fmt"

	"github.com/acasajus/menac/db"
)

type Organization struct {
	db.Record

	Handle string `db:"pk"`
	Groups []string
}

func NewOrg(db db.DB) *Organization {
	return db.LinkRecordToDB(&Organization{}).(*Organization)
}

func (o *Organization) Create() error {
	return o.GetDB().CreateNewRecord(o)
}

func (g *Organization) Store() error {
	return g.GetDB().ReplaceRecord(g)
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

func (o *Organization) CreateGroup(name string) (*Group, error) {
	for _, n := range o.Groups {
		if n == name {
			return nil, errors.New("Group already exists")
		}
	}
	g := &Group{Name: name, Organization: o.Handle}
	o.GetDB().LinkRecordToDB(g)
	if err := o.GetDB().CreateNewRecord(g); err != nil {
		return nil, err
	}
	if o.Groups == nil {
		o.Groups = make([]string, 0, 1)
	}
	o.Groups = append(o.Groups, name)
	if err := o.Store(); err != nil {
		o.GetDB().DeleteRecord(g)
		return nil, err
	}
	return g, nil
}

func (o *Organization) GetGroup(name string) (*Group, error) {
	g := &Group{Name: name, Organization: o.Handle}
	if err := o.GetDB().GetRecord(g.GetPrimaryKey(), g); err != nil {
		return nil, err
	}
	return g, nil
}

func (o *Organization) GetUser(handle string) (*User, error) {
	u := &User{Handle: handle, Organization: o.Handle}
	if err := o.GetDB().GetRecord(u.GetPrimaryKey(), u); err != nil {
		return nil, err
	}
	return u, nil
}

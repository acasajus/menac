package ostore

import (
	"errors"
	"fmt"
	"time"

	"github.com/acasajus/menac/db"
)

type StoredObject struct {
	db.Record

	User         string
	Group        string
	Organization string
	StoreName    string
	Expiration   time.Time
	Type         string
	Hash         string
	Metadata     map[string]string

	store *ObjectStore
}

func (so *StoredObject) Create() error {
	return so.GetDB().CreateNewRecord(so)
}

func (so *StoredObject) Store() error {
	return so.GetDB().ReplaceRecord(so)
}

func (so *StoredObject) Validate() error {
	if len(so.Organization) == 0 {
		return errors.New("Empty organization handle")
	}
	if len(so.User) == 0 {
		return errors.New("Empty user")
	}
	if len(so.Group) == 0 {
		return errors.New("Empty group")
	}
	if len(so.StoreName) == 0 {
		return errors.New("Empty store name")
	}
	if len(so.Hash) == 0 {
		return errors.New("Empty hash")
	}
	return nil
}

func (so *StoredObject) getPath() string {
	return fmt.Sprintf("%s/%s", so.Type, so.Hash)
}

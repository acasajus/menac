package ostore

import (
	"errors"

	"github.com/acasajus/menac/db"
	"github.com/acasajus/menac/registry"
)

const (
	STORE_TYPE_SWIFT = iota
)

type ObjectStore struct {
	db.Record

	Organization string
	Name         string
	Type         int
	data         map[string]string
}

func NewObjectStore(o *registry.Organization) *ObjectStore {
	return o.GetDB().LinkRecordToDB(&ObjectStore{}).(*ObjectStore)
}

func (so *ObjectStore) Create() error {
	return so.GetDB().CreateNewRecord(so)
}

func (so *ObjectStore) Store() error {
	return so.GetDB().ReplaceRecord(so)
}

func (so *ObjectStore) Validate() error {
	if len(so.Organization) == 0 {
		return errors.New("Empty organization handle")
	}
	if len(so.Name) == 0 {
		return errors.New("Empty store name")
	}
	return nil
}

func (os *ObjectStore) NewObject() *StoredObject {
	return os.GetDB().LinkRecordToDB(&StoredObject{
		Organization: os.Organization,
		StoreName:    os.Name,
	}).(*StoredObject)
}
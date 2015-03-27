package wms

import (
	"errors"

	"github.com/acasajus/menac/db"
	"github.com/acasajus/menac/registry"
)

type Job struct {
	db.Record

	Organization string
	User         string
	Group        string
}

func NewJob(o *registry.Organization) *Job {
	return o.GetDB().LinkRecordToDB(&Job{}).(*Job)
}

func (j *Job) Create() error {
	return j.GetDB().CreateNewRecord(j)
}

func (j *Job) Store() error {
	return j.GetDB().ReplaceRecord(j)
}

func (j *Job) Validate() error {
	if len(j.Organization) == 0 {
		return errors.New("Empty organization handle")
	}
	if len(j.User) == 0 {
		return errors.New("Empty user")
	}
	if len(j.Group) == 0 {
		return errors.New("Empty group")
	}
	return nil
}

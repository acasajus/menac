package ostore

import (
	"errors"
	"io"
)

var (
	ErrHashMismatch  = errors.New("Stored Object's expected hash did not match real hash")
	ErrNotExists     = errors.New("Object does not exist")
	ErrAlreadyExists = errors.New("Object already exists")
)

type Store interface {
	Put(*StoredObject, io.Reader, int64) error
	Get(*StoredObject, io.Writer) error
	Delete(*StoredObject) error
}

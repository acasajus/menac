package ostore

import (
	"errors"
	"io"
	"time"

	"github.com/ncw/swift"
)

type SwiftStore struct {
	conn          *swift.Connection
	containerName string
}

const (
	SWIFT_AUTH_URL    = "auth_url"
	SWIFT_USER_NAME   = "user"
	SWIFT_API_KEY     = "api_key"
	SWIFT_TENANT      = "tenant"
	SWIFT_TENANT_ID   = "tenant_id"
	SWIFT_STORAGE_URL = "storage_url"
	SWIFT_AUTH_TOKEN  = "auth_token"
	SWIFT_CONTAINER   = "container"
)

func (s *SwiftStore) Initialize(c map[string]string) error {
	if _, ok := c[SWIFT_CONTAINER]; !ok {
		return errors.New("No container defined")
	}
	s.conn = &swift.Connection{
		AuthUrl:    c[SWIFT_AUTH_URL],
		UserName:   c[SWIFT_USER_NAME],
		ApiKey:     c[SWIFT_API_KEY],
		Tenant:     c[SWIFT_TENANT],
		TenantId:   c[SWIFT_TENANT_ID],
		StorageUrl: c[SWIFT_STORAGE_URL],
		AuthToken:  c[SWIFT_AUTH_TOKEN],
	}

	if s.conn.Authenticated() {
		return nil
	}
	if err := s.conn.Authenticate(); err != nil {
		return err
	}
	if err := s.conn.ContainerCreate(c[SWIFT_CONTAINER], nil); err != nil {
		return err
	}
	s.containerName = c[SWIFT_CONTAINER]
	return nil
}

func (s *SwiftStore) Put(so *StoredObject, data io.Reader, length int64) error {
	_, _, err := s.conn.Object(s.containerName, so.getPath())
	switch err {
	case swift.ObjectNotFound:
		break
	case nil:
		return nil
	default:
		return err
	}
	headers := map[string]string{}
	if so.Expiration.After(time.Now()) {
		headers["X-Expiration"] = so.Expiration.Format(time.RFC3339)
	}
	for k, v := range so.Metadata {
		headers["X-Object-"+k] = v
	}
	hr := NewHashReader(data)
	_, err = s.conn.ObjectPut(s.containerName, so.getPath(), hr, true, "", "application/octet-stream", swift.Headers(headers))
	if err != nil {
		return err
	}
	if hr.HexDigest() != so.Hash {
		s.Delete(so)
		return ErrHashMismatch
	}
	return nil
}

func (s *SwiftStore) Get(so *StoredObject, data io.Writer) error {
	_, err := s.conn.ObjectGet(s.containerName, so.getPath(), data, true, nil)
	if err == swift.ObjectNotFound {
		return ErrNotExists
	}
	return err
}

func (s *SwiftStore) Delete(so *StoredObject) error {
	err := s.conn.ObjectDelete(s.containerName, so.getPath())
	if err == swift.ObjectNotFound {
		return nil
	}
	return err
}

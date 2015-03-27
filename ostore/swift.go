package ostore

import (
	"io"

	"github.com/ncw/swift"
)

type Swift struct {
	conn *swift.Connection
}

const (
	SWIFT_AUTH_URL    = "auth_url"
	SWIFT_USER_NAME   = "user"
	SWIFT_API_KEY     = "api_key"
	SWIFT_TENANT      = "tenant"
	SWIFT_TENANT_ID   = "tenant_id"
	SWIFT_STORAGE_URL = "storage_url"
	SWIFT_AUTH_TOKEN  = "auth_token"
)

func (s *Swift) Initialize(c map[string]string) error {
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
	return nil
}

func (s *Swift) Put(so *StoredObject, data io.Reader) error {
	_, h, err := s.conn.Object(so.Type, so.Hash)
	switch err {
	case swift.ObjectNotFound:
		break
	case nil:
		for k, v := range h {
			so.Metadata["STORE-"+k] = v
		}
		return nil
	default:
		return err
	}
	h, err = s.conn.ObjectPut(so.Type, so.Hash, data, true, "", "application/octet-stream", swift.Headers(so.Metadata))
	if err != nil {
		return err
	}
	for k, v := range h {
		so.Metadata["STORE-"+k] = v
	}
	return nil
}

func (s *Swift) Get(so *StoredObject, data io.Writer) error {
	h, err := s.conn.ObjectGet(so.Type, so.Hash, data, true, swift.Headers(so.Metadata))
	if err != nil {
		return err
	}
	for k, v := range h {
		so.Metadata["STORE-"+k] = v
	}
	return nil
}

func (s *Swift) Delete(so *StoredObject) error {
	err := s.conn.ObjectDelete(so.Type, so.Hash)
	if err == swift.ObjectNotFound {
		return nil
	}
	return err
}

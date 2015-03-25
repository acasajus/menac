package registry

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/acasajus/menac/db"
	"golang.org/x/crypto/scrypt"
)

const (
	PW_SALT_LEN = 32
	PW_KEY_KEN  = 32
)

type User struct {
	db.Record

	Handle       string
	Email        []string
	Name         string
	Password     []byte
	Organization string `db:"indexed"`
}

func (u *User) GetPrimaryKey() []byte {
	return []byte(fmt.Sprintf("%s:%s", u.Organization, u.Handle))
}

func (u *User) Validate() error {
	if len(u.Handle) == 0 {
		return errors.New("Empty user handle")
	}
	if len(u.Name) == 0 {
		return errors.New("User name is empty")
	}
	if len(u.Password) == 0 {
		return errors.New("Password is empty")
	}
	if len(u.Organization) == 0 {
		return errors.New("Organization is empty")
	}
	if len(u.Email) == 0 {
		return errors.New("Email is empty")
	}
	return nil
}

func (u *User) Create() error {
	return u.GetDB().CreateNewRecord(u)
}

func (u *User) SetPassword(p string) {
	data := make([]byte, PW_SALT_LEN+PW_KEY_KEN)
	if c, err := rand.Read(data[:PW_SALT_LEN]); err != nil {
		panic(err)
	} else if c != 32 {
		panic("Didn't read 32 chars!")
	}
	if h, err := scrypt.Key([]byte(p), data[:PW_SALT_LEN], 32768, 16, 1, PW_KEY_KEN); err != nil {
		panic(err)
	} else if c := copy(data[PW_SALT_LEN:], h); c != PW_KEY_KEN {
		panic("Didn't copy ALL PW KEY")
	}
	u.Password = data
}

func (u *User) CheckPassword(p string) bool {
	h, err := scrypt.Key([]byte(p), u.Password[:PW_SALT_LEN], 32768, 16, 1, PW_KEY_KEN)
	if err != nil {
		panic(err)
	}
	return bytes.Equal(h, u.Password[PW_SALT_LEN:])
}

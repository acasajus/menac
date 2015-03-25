package registry

import (
	"errors"
	"fmt"

	"github.com/acasajus/menac/db"
)

type Group struct {
	db.Record

	Name         string
	Users        []string
	Share        int
	Properties   []string
	Organization string

	org *Organization `db:"-"`
}

func (g *Group) GetPrimaryKey() []byte {
	return []byte(fmt.Sprintf("%s:%s", g.Organization, g.Name))
}

func (g *Group) Validate() error {
	if len(g.Name) == 0 {
		return errors.New("Name is empty")
	}
	if len(g.Organization) == 0 {
		return errors.New("Organization is empty")
	}
	return nil
}

func (g *Group) AddUsers(users ...string) {
	if len(users) == 0 {
		return
	}
	if g.Users == nil {
		g.Users = make([]string, 0, 1)
	}
	for _, u := range users {
		found := false
		for _, ku := range g.Users {
			if u == ku {
				found = true
				break
			}
		}
		if found {
			continue
		}
		g.Users = append(g.Users, u)
	}
}

func (g *Group) RemoveUsers(users ...string) {
	if len(users) == 0 {
		return
	}
	if g.Users == nil {
		return
	}
	for _, u := range users {
		for i, ku := range g.Users {
			if u == ku {
				g.Users = append(g.Users[:i], g.Users[i+1:]...)
				break
			}
		}
	}
}

func (g *Group) AddProperties(props ...string) {
	if len(props) == 0 {
		return
	}
	if g.Properties == nil {
		g.Properties = make([]string, 0, 1)
	}
	for _, u := range props {
		found := false
		for _, ku := range g.Properties {
			if u == ku {
				found = true
				break
			}
		}
		if found {
			continue
		}
		g.Properties = append(g.Properties, u)
	}
}

func (g *Group) RemoveProperties(props ...string) {
	if len(props) == 0 {
		return
	}
	if g.Properties == nil {
		return
	}
	for _, u := range props {
		for i, ku := range g.Properties {
			if u == ku {
				g.Properties = append(g.Properties[:i], g.Properties[i+1:]...)
				break
			}
		}
	}
}

func (g *Group) Store() error {
	return g.GetDB().ReplaceRecord(g)
}
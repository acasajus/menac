package registry

import "testing"

func TestAddRemoveUserFromGroup(t *testing.T) {
	o := getDummyOrg()
	g, err := o.CreateGroup("testgroup")
	if err != nil {
		t.Fatalf("Could not create group: %s", err)
	}
	users := []string{"u1", "u2", "u3"}
	g.AddUsers(users...)
	if len(g.Users) != len(users) {
		t.Fatal("Not all users were added")
	}
	g.AddUsers(users...)
	if len(g.Users) != len(users) {
		t.Fatal("A user was inserted twice")
	}
	for i := 0; i < len(g.Users); i++ {
		current := g.Users[i]
		g.RemoveUsers(current)
		for _, u := range g.Users {
			if u == current {
				t.Fatal("User was not removed")
			}
		}
		g.AddUsers(current)
	}
	if err := g.Store(); err != nil {
		t.Fatalf("Could not save group: %s", err)
	}
}

func TestAddRemoveProperties(t *testing.T) {
	o := getDummyOrg()
	g, err := o.CreateGroup("testgroup")
	if err != nil {
		t.Fatalf("Could not create group: %s", err)
	}
	props := []string{"u1", "u2", "u3"}
	g.AddProperties(props...)
	if len(g.Properties) != len(props) {
		t.Fatal("Not all props were added")
	}
	g.AddProperties(props...)
	if len(g.Properties) != len(props) {
		t.Fatal("A prop was inserted twice")
	}
	for i := 0; i < len(g.Properties); i++ {
		current := g.Properties[i]
		g.RemoveProperties(current)
		for _, u := range g.Properties {
			if u == current {
				t.Fatal("User was not removed")
			}
		}
		g.AddProperties(current)
	}
	if err := g.Store(); err != nil {
		t.Fatalf("Could not save group: %s", err)
	}
}

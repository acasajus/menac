package web

import "github.com/go-martini/martini"

type Web struct {
	router martini.Router
	martni *martini.Martini
}

func NewWeb() *Web {
	r := martini.NewRouter()
	m := martini.New()
	m.MapTo(r, (*martini.Routes)(nil))
	return &Web{r, m}
}
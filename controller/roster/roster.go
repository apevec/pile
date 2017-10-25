// Package roster
package roster

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/arapov/pile/controller/ldapxrest"
	"github.com/arapov/pile/lib/flight"

	"github.com/blue-jay/core/router"
)

var (
	uri = "/roster"
)

// Load the routes.
func Load() {
	//c := router.Chain(acl.DisallowAnon)
	router.Get(uri, Index) //, c...)
	router.Get("/ping", Ping)

	router.Get(uri+"/v2/groups", GetGroups2)
	router.Get(uri+"/v2/groups/:group/size", GetGroupSize)
	router.Get(uri+"/v2/groups/:group/head", GetGroupHead)
}

// Index displays the items.
func Index(w http.ResponseWriter, r *http.Request) {
	c := flight.Context(w, r)

	v := c.View.New("roster/index")
	v.Render(w, r)
}

func GetGroupHead(w http.ResponseWriter, r *http.Request) {
	c := flight.Context(w, r)
	w.Header().Set("Content-Type", "application/json")

	group := c.Param("group")
	head, _ := ldapxrest.GetGroupHead(c.LDAP, group)
	js, _ := json.Marshal(head)

	w.Write(js)
}

func GetGroupSize(w http.ResponseWriter, r *http.Request) {
	c := flight.Context(w, r)
	w.Header().Set("Content-Type", "application/json")

	group := c.Param("group")
	size, _ := ldapxrest.GetGroupSize(c.LDAP, group)
	js, _ := json.Marshal(size)

	w.Write(js)
}

func GetGroups2(w http.ResponseWriter, r *http.Request) {
	c := flight.Context(w, r)
	w.Header().Set("Content-Type", "application/json")

	groups, _ := ldapxrest.GetGroups(c.LDAP)
	js, _ := json.Marshal(groups)

	w.Write(js)
}

func Ping(w http.ResponseWriter, r *http.Request) {
	c := flight.Context(w, r)
	w.Header().Set("Content-Type", "application/json")

	pong, err := ldapxrest.Ping(c.LDAP)
	if err != nil {
		log.Println(err)
	}
	js, _ := json.Marshal(pong)

	w.Write(js)
}

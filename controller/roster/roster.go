// Package roster
package roster

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/arapov/pile/controller/ldapxrest"
	"github.com/arapov/pile/lib/flight"
	"github.com/arapov/pile/model/roster"

	"github.com/blue-jay/core/router"
)

var (
	uri = "/roster"
)

// Load the routes.
func Load() {
	//c := router.Chain(acl.DisallowAnon)
	router.Get(uri, Index) //, c...)
	router.Get(uri+"/v1/groups", GetGroups)
	router.Get(uri+"/v2/groups", GetGroups2)
	router.Get(uri+"/v1/members/:group", GetMembers)
	router.Get("/ping", Ping)
}

// Index displays the items.
func Index(w http.ResponseWriter, r *http.Request) {
	c := flight.Context(w, r)

	v := c.View.New("roster/index")
	v.Render(w, r)
}

func GetMembers(w http.ResponseWriter, r *http.Request) {
	c := flight.Context(w, r)
	w.Header().Set("Content-Type", "application/json")

	group := c.Param("group")
	members, _ := roster.GetMembers(c.LDAP, group)
	js, _ := json.Marshal(members)

	w.Write(js)
}

func GetGroups(w http.ResponseWriter, r *http.Request) {
	c := flight.Context(w, r)
	w.Header().Set("Content-Type", "application/json")

	groups, _ := roster.GetGroups(c.LDAP)
	js, _ := json.Marshal(groups)

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

	pong, err := roster.Ping(c.LDAP)
	if err != nil {
		log.Println(err)
	}
	js, _ := json.Marshal(pong)

	w.Write(js)
}

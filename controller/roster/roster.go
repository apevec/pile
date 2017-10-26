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

	router.Get(uri+"/v2/people/:uid/tz", GetTimezoneInfo)

	router.Get(uri+"/v2/groups", GetGroups)
	router.Get(uri+"/v2/groups/:group", GetGroup)
	router.Get(uri+"/v2/groups/:group/size", GetGroupSize)
	router.Get(uri+"/v2/groups/:group/head", GetGroupHead)
	router.Get(uri+"/v2/groups/:group/links", GetGroupLinks)
	router.Get(uri+"/v2/groups/:group/members", GetGroupMembers)
}

func GetGroup(w http.ResponseWriter, r *http.Request) {
	c := flight.Context(w, r)
	w.Header().Set("Content-Type", "application/json")

	type info struct {
		Links map[string]string
		Head  map[string][]map[string]string
		Size  map[string]int
	}

	group := c.Param("group")
	links, _ := ldapxrest.GetGroupLinks(c.LDAP, group)
	head, _ := ldapxrest.GetGroupHead(c.LDAP, group)
	size, _ := ldapxrest.GetGroupSize(c.LDAP, group)

	groupInfo := &info{
		Links: links,
		Head:  head,
		Size:  size,
	}

	js, _ := json.Marshal(groupInfo)
	w.Write(js)
}

func GetTimezoneInfo(w http.ResponseWriter, r *http.Request) {
	c := flight.Context(w, r)
	w.Header().Set("Content-Type", "application/json")

	uid := c.Param("uid")
	tzinfo, _ := ldapxrest.GetTimezoneInfo(c.LDAP, uid)
	js, _ := json.Marshal(tzinfo)

	w.Write(js)
}

func GetGroupMembers(w http.ResponseWriter, r *http.Request) {
	c := flight.Context(w, r)
	w.Header().Set("Content-Type", "application/json")

	group := c.Param("group")
	members, _ := ldapxrest.GetGroupMembers(c.LDAP, group)
	js, _ := json.Marshal(members)

	w.Write(js)
}

func GetGroupLinks(w http.ResponseWriter, r *http.Request) {
	c := flight.Context(w, r)
	w.Header().Set("Content-Type", "application/json")

	group := c.Param("group")
	links, _ := ldapxrest.GetGroupLinks(c.LDAP, group)
	js, _ := json.Marshal(links)

	w.Write(js)
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

func GetGroups(w http.ResponseWriter, r *http.Request) {
	c := flight.Context(w, r)
	w.Header().Set("Content-Type", "application/json")

	groups, _ := ldapxrest.GetGroups(c.LDAP)
	js, _ := json.Marshal(groups)

	w.Write(js)
}

func Index(w http.ResponseWriter, r *http.Request) {
	c := flight.Context(w, r)

	v := c.View.New("roster/index")
	v.Render(w, r)
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

package ldapxrest

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/arapov/pile/lib/flight"
	"github.com/arapov/pile/model/ldapxrest"
	"github.com/blue-jay/core/router"
)

var (
	uri = "/api"
)

func Load() {
	router.Get("/ping", Ping)

	router.Get(uri+"/v2/people/:uid/tz", GetTimezoneInfo)
	router.Get(uri+"/v2/groups/heads", GetHeads)

	router.Get(uri+"/v2/groups", GetGroups)
	router.Get(uri+"/v2/groups/:group", GetGroups)
	router.Get(uri+"/v2/groups/:group/info", GetGroupInfo)
	router.Get(uri+"/v2/groups/:group/size", GetGroupSize)
	router.Get(uri+"/v2/groups/:group/head", GetGroupHead)
	router.Get(uri+"/v2/groups/:group/links", GetGroupLinks)
	router.Get(uri+"/v2/groups/:group/members", GetGroupMembers)
}

func GetHeads(w http.ResponseWriter, r *http.Request) {
	c := flight.Context(w, r)
	w.Header().Set("Content-Type", "application/json")

	var heads = make(map[string]map[string]string)

	roles, _ := ldapxrest.GetRoles(c.LDAP)
	for role := range roles {
		headPeople, _ := ldapxrest.GetPeople(c.LDAP, roles[role].Members)

		for head, name := range headPeople {
			var info = make(map[string]string)
			info["uid"] = head
			info["name"] = name
			info["role"] = roles[role].Name
			info["group"] = "tbd"

			heads[head] = info
		}
	}

	groups, _ := ldapxrest.GetGroups(c.LDAP)
	for group, groupName := range groups {
		members, _ := ldapxrest.GetGroupMembersSlice(c.LDAP, group)

		for _, member := range members {
			if _, ok := heads[member]; !ok {
				continue
			}

			// In case we have one person in more than one group
			// we clone this person with another key [9:12]
			// This is useful to have it this way, as we can
			// spot Head folks who aren't assigned to any group
			if heads[member]["group"] != "tbd" {
				var newinfo = make(map[string]string)
				for k, v := range heads[member] {
					newinfo[k] = v
					newinfo["group"] = groupName
				}
				heads[member+group[9:12]] = newinfo
				continue
			}

			heads[member]["group"] = groupName
		}
	}

	js, _ := json.Marshal(heads)
	w.Write(js)
}

func GetGroupInfo(w http.ResponseWriter, r *http.Request) {
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

	group := c.Param("group")
	groups, _ := ldapxrest.GetGroups(c.LDAP, group)
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

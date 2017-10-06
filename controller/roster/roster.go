// Package roster
package roster

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"

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
	router.Get(uri+"/v1/groups", Get)
	router.Get(uri+"/v1/members/:group", GetMembers)
	//	router.Post(uri+"/create", Store, c...)
	//	router.Get(uri+"/view/:id", Show, c...)
	//	router.Get(uri+"/edit/:id", Edit, c...)
	//	router.Patch(uri+"/edit/:id", Update, c...)
	//	router.Delete(uri+"/:id", Destroy, c...)
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

	// TODO: check whether any param is passed at all
	group := c.Param("group")
	ppl := roster.GetMembers(c.LDAP, group)
	js, _ := json.Marshal(ppl)
	w.Write(js)
}

// Get some
func Get(w http.ResponseWriter, r *http.Request) {
	c := flight.Context(w, r)
	w.Header().Set("Content-Type", "application/json")

	dfgs := roster.GetDFGroups(c.LDAP)
	// Sort dfgs by dfgs.Name
	sort.Slice(dfgs, func(i, j int) bool {
		switch strings.Compare(dfgs[i].Desc, dfgs[j].Desc) {
		case -1:
			return true
		case 1:
			return false
		}
		return dfgs[i].Desc > dfgs[j].Desc
	})
	js, _ := json.Marshal(dfgs)
	// TODO: check for errors

	w.Write(js)
}

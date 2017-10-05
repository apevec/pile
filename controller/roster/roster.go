// Package roster
package roster

import (
	"encoding/json"
	"net/http"

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
	router.Get(uri+"/get", Get)
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

// Get some
func Get(w http.ResponseWriter, r *http.Request) {
	c := flight.Context(w, r)
	w.Header().Set("Content-Type", "application/json")

	dfgs := roster.GetDFGroups(c.LDAP)
	js, _ := json.Marshal(dfgs)
	// TODO: check for errors

	w.Write(js)
}

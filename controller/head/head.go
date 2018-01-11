// Package head
package head

import (
	"net/http"

	"github.com/arapov/pile/lib/flight"

	"github.com/blue-jay/core/router"
)

var (
	uri = "/roster"
)

// Load the routes.
func Load() {
	router.Get(uri+"/head", IndexHead)
	router.Get(uri+"/all", IndexAll)
}

// Index displays the items.
func IndexHead(w http.ResponseWriter, r *http.Request) {
	c := flight.Context(w, r)

	v := c.View.New("head/index")
	v.Vars["name"] = "TC-UA-Steward"
	v.Vars["suffix"] = "heads"
	v.Render(w, r)
}

// Index displays the items.
func IndexAll(w http.ResponseWriter, r *http.Request) {
	c := flight.Context(w, r)

	v := c.View.New("head/index")
	v.Vars["name"] = "Everyone in organization"
	v.Vars["suffix"] = "all"
	v.Render(w, r)
}

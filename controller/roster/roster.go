// Package roster
package roster

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
	router.Get(uri, Index)
	router.Get(uri+"/:group", Index)
}

func Index(w http.ResponseWriter, r *http.Request) {
	c := flight.Context(w, r)

	group := c.Param("group")
	v := c.View.New("roster/index")
	v.Vars["group"] = group

	v.Render(w, r)
}

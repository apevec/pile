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
	//c := router.Chain(acl.DisallowAnon)
	router.Get(uri, Index) //, c...)
}

func Index(w http.ResponseWriter, r *http.Request) {
	c := flight.Context(w, r)

	v := c.View.New("roster/index")
	v.Render(w, r)
}

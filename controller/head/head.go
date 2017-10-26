// Package head
package head

import (
	"net/http"

	"github.com/arapov/pile/lib/flight"

	"github.com/blue-jay/core/router"
)

var (
	uri = "/roster/head"
)

// Load the routes.
func Load() {
	router.Get(uri, Index)
}

// Index displays the items.
func Index(w http.ResponseWriter, r *http.Request) {
	c := flight.Context(w, r)

	v := c.View.New("head/index")
	v.Render(w, r)
}

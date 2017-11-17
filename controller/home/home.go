// Package home displays the Home page.
package home

import (
	"net/http"

	"github.com/arapov/pile/lib/flight"

	"github.com/blue-jay/core/router"
)

// TODO:
// - get the git specifics to the /lib

const (
	directory = "pages"
)

var (
	uri = "/"
)

// Load the routes.
func Load() {
	router.Get(uri, Index)
}

// Index displays the home page.
func Index(w http.ResponseWriter, r *http.Request) {
	c := flight.Context(w, r)

	//output := gitpages.GetPage()

	v := c.View.New("home/index")
	v.Render(w, r)
}

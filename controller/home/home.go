// Package home displays the Home page.
package home

import (
	"net/http"

	"github.com/arapov/pile/lib/flight"
	"github.com/arapov/pile/model/gitpages"

	"github.com/blue-jay/core/router"
)

var (
	uri = "/"
)

// Load the routes.
func Load() {
	router.Get(uri, Index)
	router.Get(uri+"edit", Edit)
}

// Index displays the home page.
func Index(w http.ResponseWriter, r *http.Request) {
	c := flight.Context(w, r)

	page, modified := gitpages.GetPage()

	v := c.View.New("home/index")
	v.Vars["page"] = page
	v.Vars["modified"] = modified
	v.Render(w, r)
}

func Edit(w http.ResponseWriter, r *http.Request) {
	c := flight.Context(w, r)

	v := c.View.New("home/edit")
	v.Render(w, r)
}

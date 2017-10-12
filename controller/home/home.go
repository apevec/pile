// Package home displays the Home page.
package home

import (
	"net/http"
	"os"

	"github.com/arapov/pile/lib/flight"
	"github.com/arapov/pile/model/gitwiki"

	"github.com/blue-jay/core/router"
)

// TODO:
// - get the git specifics to the /lib

const (
	directory = "wiki"
)

var (
	uri = "/wiki"
)

// Load the routes.
func Load() {
	router.Get("/", Redirect)
	router.Get(uri, Index)
	router.Get(uri+"/create", Clone)
	router.Post(uri+"/create", Clone)
}

// exists returns whether the given file or directory exists or not
func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}

// Clone - clones git repo, where the wiki pages are stored
func Clone(w http.ResponseWriter, r *http.Request) {
	c := flight.Context(w, r)

	v := c.View.New("home/create")
	v.Render(w, r)
}

// Index displays the home page.
func Index(w http.ResponseWriter, r *http.Request) {
	c := flight.Context(w, r)

	if !exists(directory) {
		c.Redirect(uri + "/create")
	}

	output := gitwiki.GetPage()

	v := c.View.New("home/index")
	v.Vars["data"] = output
	v.Render(w, r)
}

// Redirect - redirects to uri, we don't have root
func Redirect(w http.ResponseWriter, r *http.Request) {
	c := flight.Context(w, r)

	c.Redirect(uri)
}

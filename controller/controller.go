// Package controller loads the routes for each of the controllers.
package controller

import (
	"github.com/arapov/pile/controller/about"
	"github.com/arapov/pile/controller/head"
	"github.com/arapov/pile/controller/home"
	"github.com/arapov/pile/controller/ldapxrest"
	"github.com/arapov/pile/controller/roster"
	"github.com/arapov/pile/controller/static"
	"github.com/arapov/pile/controller/status"
)

// LoadRoutes loads the routes for each of the controllers.
func LoadRoutes() {
	about.Load()
	home.Load()
	static.Load()
	status.Load()
	roster.Load()
	head.Load()
	ldapxrest.Load()
}

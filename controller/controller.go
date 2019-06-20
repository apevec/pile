// Package controller loads the routes for each of the controllers.
package controller

import (
	"github.com/apevec/pile/controller/about"
	"github.com/apevec/pile/controller/head"
	"github.com/apevec/pile/controller/home"
	"github.com/apevec/pile/controller/ldapxrest"
	"github.com/apevec/pile/controller/roster"
	"github.com/apevec/pile/controller/static"
	"github.com/apevec/pile/controller/status"
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

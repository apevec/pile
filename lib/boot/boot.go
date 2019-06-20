// Package boot handles the initialization of the web components.
package boot

import (
	"log"

	"github.com/apevec/pile/controller"
	"github.com/apevec/pile/lib/env"
	"github.com/apevec/pile/lib/flight"
	"github.com/apevec/pile/viewfunc/link"
	"github.com/apevec/pile/viewfunc/noescape"
	"github.com/apevec/pile/viewfunc/prettytime"
	"github.com/apevec/pile/viewmodify/authlevel"
	"github.com/apevec/pile/viewmodify/flash"
	"github.com/apevec/pile/viewmodify/uri"

	"github.com/blue-jay/core/form"
	"github.com/blue-jay/core/pagination"
	"github.com/blue-jay/core/xsrf"
)

// RegisterServices sets up all the web components.
func RegisterServices(config *env.Info) {
	// Set up the session cookie store
	err := config.Session.SetupConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Connect to the MySQL database
	mysqlDB, _ := config.MySQL.Connect(true)

	// Connect to LDAP
	ldapClient, err := config.LDAP.Dial()
	if err != nil {
		log.Panic(err)
	}

	// Load the controller routes
	controller.LoadRoutes()

	// Set up the views
	config.View.SetTemplates(config.Template.Root, config.Template.Children)

	// Set up the functions for the views
	config.View.SetFuncMaps(
		config.Asset.Map(config.View.BaseURI),
		link.Map(config.View.BaseURI),
		noescape.Map(),
		prettytime.Map(),
		form.Map(),
		pagination.Map(),
	)

	// Set up the variables and modifiers for the views
	config.View.SetModifiers(
		authlevel.Modify,
		uri.Modify,
		xsrf.Token,
		flash.Modify,
	)

	// Store the variables in flight
	flight.StoreConfig(*config)

	// Store LDAP connection in flight
	flight.StoreLDAP(ldapClient)

	// Store the database connection in flight
	flight.StoreDB(mysqlDB)

	// Store the csrf information
	flight.StoreXsrf(xsrf.Info{
		AuthKey: config.Session.CSRFKey,
		Secure:  config.Session.Options.Secure,
	})
}

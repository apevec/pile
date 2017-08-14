// Package home displays the Home page.
package home

import (
	"encoding/json"
	"net/http"

	"github.com/arapov/pile/lib/flight"
	"github.com/arapov/pile/model/ldap"

	"github.com/blue-jay/core/router"
)

// Load the routes.
func Load() {
	router.Get("/", Index)
}

// Index displays the home page.
func Index(w http.ResponseWriter, r *http.Request) {
	c := flight.Context(w, r)

	people, dfgs, err := ldap.GetAll(c.LDAP)
	if err != nil {
		c.FlashError(err)
		return
	}

	jdfgs, err := json.Marshal(dfgs)
	if err != nil {
		c.FlashError(err)
		jdfgs = []byte(`{}`)
	}
	jpeople, err := json.Marshal(people)
	if err != nil {
		c.FlashError(err)
		jpeople = []byte(`{}`)
	}

	v := c.View.New("home/index")
	if c.Sess.Values["id"] != nil {
		v.Vars["first_name"] = c.Sess.Values["first_name"]
	}
	v.Vars["dfgs"] = jdfgs
	v.Vars["people"] = jpeople
	v.Render(w, r)
}

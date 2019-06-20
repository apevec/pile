// Package home displays the Home page.
package home

import (
	"net/http"

	"github.com/apevec/pile/lib/flight"
	"github.com/apevec/pile/model/gitpages"

	"github.com/blue-jay/core/router"
)

var (
	uri = "/"
)

// Load the routes.
func Load() {
	router.Get(uri, Index)
	router.Patch(uri+"edit", Update)
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

	page, _ := gitpages.GetPageRaw()

	v := c.View.New("home/edit")
	c.Repopulate(v.Vars, "change")
	c.Repopulate(v.Vars, "page")
	if v.Vars["page"] == nil {
		v.Vars["page"] = string(page)
	}
	v.Render(w, r)
}

func Update(w http.ResponseWriter, r *http.Request) {
	c := flight.Context(w, r)

	if !c.FormValid("page") {
		Edit(w, r)
		return
	}
	if !c.FormValid("change") {
		Edit(w, r)
		return
	}

	err := gitpages.Update(r.FormValue("page"), r.FormValue("change"))
	if err != nil {
		c.FlashErrorGeneric(err)
		Edit(w, r)
		return
	}

	c.FlashSuccess("Page updated.")
	c.Redirect(uri)
}

// Package head
package head

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/arapov/pile/lib/flight"
	"github.com/arapov/pile/model/head"

	"github.com/blue-jay/core/router"
)

var (
	uri = "/roster/head"
)

// Load the routes.
func Load() {
	router.Get(uri, Index)
	router.Get("/roster/v1/head", GetHead)
}

func GetHead(w http.ResponseWriter, r *http.Request) {
	c := flight.Context(w, r)
	w.Header().Set("Content-Type", "application/json")

	head, err := head.GetHead(c.LDAP)
	if err != nil {
		log.Println(err)
	}
	js, _ := json.Marshal(head)

	w.Write(js)
}

// Index displays the items.
func Index(w http.ResponseWriter, r *http.Request) {
	c := flight.Context(w, r)

	v := c.View.New("head/index")
	v.Render(w, r)
}

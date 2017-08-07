// Package ldap
package ldap

import (
	"fmt"
	"log"

	"github.com/arapov/pile/lib/flight"
)

// Item defines the model.
type Item struct {
}

// Connection is an interface for making queries.
type Connection interface {
	//	Exec(query string, args ...interface{}) (sql.Result, error)
	//	Get(dest interface{}, query string, args ...interface{}) error
	//	Select(dest interface{}, query string, args ...interface{}) error
}

// ByID gets an item by ID.
func ByID(db Connection, ID string, userID string) (Item, error) {
	var result Item
	var err error

	log.Println("Implementing")

	result := "item"

	return result, err
}

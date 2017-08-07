// Package ldap
package ldap

import (
	"fmt"
	"log"

	"gopkg.in/ldap.v2"
)

// Info holds the config.
type Info struct {
	Hostname string
	Port     int
}

// ByID gets an item by ID.
func (c Info) Search() (string, error) {
	var err error

	log.Println("Implementing")

	l, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", c.Hostname, c.Port))
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	searchRequest := ldap.NewSearchRequest(
		"ou=users,dc=redhat,dc=com", // The base dn to search
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectClass=organizationalPerson))", // The filter to apply
		[]string{"dn", "cn"},                    // A list attributes to retrieve
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		log.Fatal(err)
	}

	for _, entry := range sr.Entries {
		fmt.Printf("%s: %v\n", entry.DN, entry.GetAttributeValue("cn"))
	}

	result := "something works"

	return result, err
}

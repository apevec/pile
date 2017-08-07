// Package ldap
package ldap

import (
	"log"

	"gopkg.in/ldap.v2"
)

type Item struct {
	dn string
	cn string
}

type Connection interface {
	Search(*ldap.SearchRequest) (*ldap.SearchResult, error)
}

func Get(ldapc Connection) ([]Item, error) {
	var result []Item

	searchRequest := ldap.NewSearchRequest(
		"ou=users,dc=redhat,dc=com", // The base dn to search
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectClass=organizationalPerson))", // The filter to apply
		[]string{"dn", "cn"},                    // A list attributes to retrieve
		nil,
	)

	sr, err := ldapc.Search(searchRequest)
	if err != nil {
		// TODO: Must not .Fatal
		log.Fatal(err)
	}

	for _, entry := range sr.Entries {
		result = append(result, Item{entry.DN, entry.GetAttributeValue("cn")})
	}

	return result, err
}

// Package roster
package roster

import (
	"fmt"

	ldap "gopkg.in/ldap.v2"
)

// Item defines the model.
type Item struct {
}

// DFGroup defines the DFG model.
type DFGroup struct {
	Name    string
	Desc    string
	Members []string
	Backlog string
	Mission string
}

// Connection is an interface for making queries.
type Connection interface {
	Search(*ldap.SearchRequest) (*ldap.SearchResult, error)
}

// GetDFGroups - tbd
func GetDFGroups(ldapc Connection) []DFGroup {
	var dfgs []DFGroup

	sGroupRequest := ldap.NewSearchRequest(
		"ou=adhoc,ou=managedGroups,dc=redhat,dc=com",
		ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectClass=rhatGroup)(&(cn=rhos-dfg-*)(!(cn=*squad*))))",
		[]string{"cn", "description", "memberUid", "rhatGroupNotes"},
		nil,
	)

	searchResult, _ := ldapc.Search(sGroupRequest)
	// TODO: check for err

	for _, entry := range searchResult.Entries {
		name := entry.GetAttributeValue("cn")

		sSquadRequest := ldap.NewSearchRequest(
			"ou=adhoc,ou=managedGroups,dc=redhat,dc=com",
			ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false,
			fmt.Sprintf("(&(objectClass=rhatGroup)(cn=%s-*))", name),
			[]string{"cn", "description", "memberUid", "rhatGroupNotes"},
			nil,
		)

		ldapSquads, _ := ldapc.Search(sSquadRequest)
		// TODO: check for err
		if len(ldapSquads.Entries) > 0 {
			// TODO: work w/ squads
		}

		dfg := DFGroup{
			Name:    entry.GetAttributeValue("cn"),
			Desc:    entry.GetAttributeValue("description"),
			Members: entry.GetAttributeValues("memberUid"),
			Backlog: entry.GetAttributeValue("rhatGroupNotes"),
			Mission: entry.GetAttributeValue("rhatGroupNotes"),
		}
		dfgs = append(dfgs, dfg)
	}
	//fmt.Println(dfgs)
	return dfgs
}

// GetDFGroupMembers - tbd
func GetDFGroupMembers(ldapc Connection) {

}

// Package ldap
package ldap

import (
	"fmt"

	"gopkg.in/ldap.v2"
)

/* Assumptions:
- no person must belong to more than one group
  - non- UA, TC, PM
- no person may have more than one role
*/

type Item struct {
	mail  string
	title string
	cn    string
	dfg   []string
	role  []string
}

type Items map[string]*Item

type Connection interface {
	Search(*ldap.SearchRequest) (*ldap.SearchResult, error)
}

func Get(ldapc Connection) (Items, error) {
	items := make(Items)

	// TODO: Make it dedup, sane and readable

	// DFGs:
	searchRequest := ldap.NewSearchRequest(
		"ou=adhoc,ou=managedGroups,dc=redhat,dc=com",
		ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectClass=rhatGroup)(cn=rhos-dfg-*))",
		[]string{"cn", "memberUid"},
		nil,
	)

	searchResult, err := ldapc.Search(searchRequest)
	if err != nil {
		fmt.Println(err)
	}

	for _, entry := range searchResult.Entries {
		for _, member := range entry.GetAttributeValues("memberUid") {
			if _, ok := items[member]; !ok {
				items[member] = &Item{}
			}
			items[member].dfg = append(items[member].dfg, entry.GetAttributeValue("cn"))
		}
	}

	// Roles:
	searchRequest = ldap.NewSearchRequest(
		"ou=adhoc,ou=managedGroups,dc=redhat,dc=com",
		ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectClass=rhatGroup)(|(cn=rhos-ua)(cn=rhos-pm)(cn=rhos-tc)))",
		[]string{"cn", "memberUid"},
		nil,
	)

	searchResult, err = ldapc.Search(searchRequest)
	if err != nil {
		fmt.Println(err)
	}

	for _, entry := range searchResult.Entries {
		for _, member := range entry.GetAttributeValues("memberUid") {
			if _, ok := items[member]; !ok {
				items[member] = &Item{}
			}
			items[member].role = append(items[member].role, entry.GetAttributeValue("cn"))
		}
	}

	// People:
	filter := "(&(objectClass=rhatPerson)(|"
	for member := range items {
		// "(&(objectClass=rhatPerson)(|(uid=user1)(uid=user2)(uid=user3)))"
		filter = filter + fmt.Sprintf("(uid=%s)", member)
	}
	filter = filter + "))"

	searchRequest = ldap.NewSearchRequest(
		"ou=users,dc=redhat,dc=com",
		ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false,
		filter, // The filter to apply
		[]string{"uid", "mail", "title", "cn"},
		nil,
	)

	sr, err := ldapc.Search(searchRequest)
	if err != nil {
		fmt.Println(err)
	}

	var uid string
	for _, entry := range sr.Entries {
		uid = entry.GetAttributeValue("uid")
		items[uid].mail = entry.GetAttributeValue("mail")
		items[uid].title = entry.GetAttributeValue("title")
		items[uid].cn = entry.GetAttributeValue("cn")
	}

	/*
		for i := range items {
			fmt.Printf("%s :", i)
			fmt.Printf("%+v\n", items[i])
		}
	*/
	return items, err
}

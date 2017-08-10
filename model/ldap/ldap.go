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
	Mail   string
	Title  string
	Cn     string
	Mobile string
	Dfg    []string
	Role   []string
}

type Items map[string]*Item

type Connection interface {
	Search(*ldap.SearchRequest) (*ldap.SearchResult, error)
}

func removeDuplicatesUnordered(elements []string) []string {
	encountered := map[string]bool{}

	// Create a map of all unique elements.
	for v := range elements {
		encountered[elements[v]] = true
	}

	// Place all keys from the map into a slice.
	result := []string{}
	for key, _ := range encountered {
		result = append(result, key)
	}
	return result
}

func GetAll(ldapc Connection) (Items, []string, error) {
	items := make(Items)
	var dfgs []string

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
			items[member].Dfg = append(items[member].Dfg, entry.GetAttributeValue("cn"))
			dfgs = append(dfgs, entry.GetAttributeValue("cn"))
		}
	}
	dfgs = removeDuplicatesUnordered(dfgs)

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
			items[member].Role = append(items[member].Role, entry.GetAttributeValue("cn"))
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
		[]string{"uid", "mail", "title", "cn", "mobile"},
		nil,
	)

	sr, err := ldapc.Search(searchRequest)
	if err != nil {
		fmt.Println(err)
	}

	var uid string
	for _, entry := range sr.Entries {
		uid = entry.GetAttributeValue("uid")
		items[uid].Mail = entry.GetAttributeValue("mail")
		items[uid].Title = entry.GetAttributeValue("title")
		items[uid].Cn = entry.GetAttributeValue("cn")
		items[uid].Mobile = entry.GetAttributeValue("mobile")
	}

	/*
		for i := range items {
			fmt.Printf("%s :", i)
			fmt.Printf("%+v\n", items[i])
		}
	*/
	return items, dfgs, err
}

// Package roster
package roster

import (
	"fmt"
	"regexp"
	"strings"

	ldap "gopkg.in/ldap.v2"
)

// Item defines the model.
type Item struct {
}

// DFGroup defines the DFG model.
type DFGroup struct {
	Name     string
	Desc     string
	Members  []string
	Backlog  string
	Mission  string
	PMs      []Person
	Stewards []Person
	Squads   map[string]string
	SquadsSz int
}

var dfgs []DFGroup

type Person struct {
	Uid  string
	Name string
	Role string
}

var people = map[string]*Person{}

// Connection is an interface for making queries.
type Connection interface {
	Search(*ldap.SearchRequest) (*ldap.SearchResult, error)
}

func decodeNote(note string) map[string]string {
	result := make(map[string]string)

	re, _ := regexp.Compile(`pile:(\w*=[a-zA-z0-9:/.-]+)`)
	// TODO: take care of error here
	pile := re.FindAllStringSubmatch(note, -1)
	// TODO: code below is fragile, very fragile
	for i := range pile {
		kv := strings.Split(pile[i][1], "=")
		result[kv[0]] = kv[1]
	}

	return result
}

func fillRoles(ldapc Connection) {
	sRolesRequest := ldap.NewSearchRequest(
		"ou=adhoc,ou=managedGroups,dc=redhat,dc=com",
		ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectClass=rhatGroup)(|(cn=rhos-ua)(cn=rhos-pm)(cn=rhos-tc)(cn=rhos-stewards-em)(cn=rhos-stewards-qe)(cn=rhos-squad-lead)))",
		[]string{"cn", "memberUid"},
		nil,
	)
	ldapRolesPeople, _ := ldapc.Search(sRolesRequest)
	// TODO: check for err

	for _, rolesPeople := range ldapRolesPeople.Entries {
		role := rolesPeople.GetAttributeValue("cn")
		for _, person := range rolesPeople.GetAttributeValues("memberUid") {
			if _, ok := people[person]; !ok {
				people[person] = &Person{}
			}
			people[person] = &Person{}
			people[person].Uid = person
			people[person].Role = role
		}
	}
}

func fillDFGs(ldapc Connection) {
	sGroupRequest := ldap.NewSearchRequest(
		"ou=adhoc,ou=managedGroups,dc=redhat,dc=com",
		ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectClass=rhatGroup)(&(cn=rhos-dfg-*)(!(cn=*squad*))))",
		[]string{"cn", "description", "memberUid", "rhatGroupNotes"},
		nil,
	)

	ldapGroups, _ := ldapc.Search(sGroupRequest)
	// TODO: check for err

	for _, group := range ldapGroups.Entries {
		dfg := DFGroup{}
		name := group.GetAttributeValue("cn")

		sSquadRequest := ldap.NewSearchRequest(
			"ou=adhoc,ou=managedGroups,dc=redhat,dc=com",
			ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false,
			fmt.Sprintf("(&(objectClass=rhatGroup)(cn=%s-*))", name),
			[]string{"cn", "description", "memberUid", "rhatGroupNotes"},
			nil,
		)

		// rhatGroupNotes is in plain text as of 10/06
		// syntax: pile:[keyword]=[value]
		note := group.GetAttributeValue("rhatGroupNotes")
		if len(note) > 0 {
			kv := decodeNote(note)
			// TODO: check for keys availability
			dfg.Backlog = kv["backlog"]
			dfg.Mission = kv["mission"]
		}

		// Check whether Group has Squads
		// TODO: don't call this ldap search for every group
		ldapSquads, _ := ldapc.Search(sSquadRequest)
		// TODO: check for err
		if len(ldapSquads.Entries) > 0 {
			dfg.Squads = make(map[string]string)
			dfg.SquadsSz = 0
			for _, squad := range ldapSquads.Entries {
				dfg.Squads[squad.GetAttributeValue("cn")] = squad.GetAttributeValue("description")
				dfg.SquadsSz++
				// TODO: here we want to decode notes for squad specific details
			}
		}

		dfg.Name = group.GetAttributeValue("cn")
		dfg.Desc = group.GetAttributeValue("description")

		for _, member := range group.GetAttributeValues("memberUid") {
			if _, ok := people[member]; ok {
				switch people[member].Role {
				case "rhos-pm":
					dfg.PMs = append(dfg.PMs, *people[member])
				case "rhos-stewards-em":
				case "rhos-stewards-qe":
					dfg.Stewards = append(dfg.Stewards, *people[member])
				}
			}
		}

		dfgs = append(dfgs, dfg)
	}
}

// GetPeople - tbd
func GetMembers(ldapc Connection, group string) []Person {

	for _, grp := range dfgs {
		if grp.Name == group {
			return grp.PMs
		}
	}

	return []Person{}
}

// GetDFGroups - tbd
func GetDFGroups(ldapc Connection) []DFGroup {

	if len(dfgs) == 0 {
		fillRoles(ldapc)
		fillDFGs(ldapc)
	}

	//fmt.Println(dfgs)
	return dfgs
}

// GetDFGroupMembers - tbd
func GetDFGroupMembers(ldapc Connection) {
	return
}

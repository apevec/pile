// Package roster - tbd
package roster

import (
	"fmt"
	"regexp"
	"strings"

	ldap "gopkg.in/ldap.v2"
)

// Member - tbd
type Member struct {
	UID  string
	Name string
	Role string
}

// Group defines the DFG model.
type Group struct {
	Name     string
	Desc     string
	Members  []string
	Backlog  string
	Mission  string
	PMs      []Member
	Stewards []Member
	Squads   map[string]string
	SquadsSz int
}

var groups []Group
var people = map[string]*Member{}

// Connection is an interface for making queries.
type Connection interface {
	Search(*ldap.SearchRequest) (*ldap.SearchResult, error)
}

// decodeNote - returns kv
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
				people[person] = &Member{}
			}
			people[person] = &Member{}
			people[person].UID = person
			people[person].Role = role
		}
	}
}

func fillGroups(ldapc Connection) {
	sGroupRequest := ldap.NewSearchRequest(
		"ou=adhoc,ou=managedGroups,dc=redhat,dc=com",
		ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectClass=rhatGroup)(&(cn=rhos-dfg-*)(!(cn=*squad*))))",
		[]string{"cn", "description", "memberUid", "rhatGroupNotes"},
		nil,
	)

	ldapGroups, _ := ldapc.Search(sGroupRequest)
	// TODO: check for err

	for _, ldapGroup := range ldapGroups.Entries {
		group := Group{}
		name := ldapGroup.GetAttributeValue("cn")

		sSquadRequest := ldap.NewSearchRequest(
			"ou=adhoc,ou=managedGroups,dc=redhat,dc=com",
			ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false,
			fmt.Sprintf("(&(objectClass=rhatGroup)(cn=%s-*))", name),
			[]string{"cn", "description", "memberUid", "rhatGroupNotes"},
			nil,
		)

		// rhatGroupNotes is in plain text as of 10/06
		// syntax: pile:[keyword]=[value]
		note := ldapGroup.GetAttributeValue("rhatGroupNotes")
		if len(note) > 0 {
			kv := decodeNote(note)
			// TODO: check for keys availability
			group.Backlog = kv["backlog"]
			group.Mission = kv["mission"]
		}

		// Check whether Group has Squads
		// TODO: don't call this ldap search for every group
		ldapSquads, _ := ldapc.Search(sSquadRequest)
		// TODO: check for err
		if len(ldapSquads.Entries) > 0 {
			group.Squads = make(map[string]string)
			group.SquadsSz = 0
			for _, ldapSquad := range ldapSquads.Entries {
				group.Squads[ldapSquad.GetAttributeValue("cn")] = ldapSquad.GetAttributeValue("description")
				group.SquadsSz++
				// TODO: here we want to decode notes for squad specific details
			}
		}

		group.Name = ldapGroup.GetAttributeValue("cn")
		group.Desc = ldapGroup.GetAttributeValue("description")

		for _, member := range ldapGroup.GetAttributeValues("memberUid") {
			if _, ok := people[member]; ok {
				switch people[member].Role {
				case "rhos-pm":
					group.PMs = append(group.PMs, *people[member])
				case "rhos-stewards-em":
				case "rhos-stewards-qe":
					group.Stewards = append(group.Stewards, *people[member])
				}
			}
		}

		groups = append(groups, group)
	}
}

// GetMembers - tbd
func GetMembers(ldapc Connection, group string) []Member {

	for _, grp := range groups {
		if grp.Name == group {
			return grp.PMs
		}
	}

	return []Member{}
}

// GetGroups - tbd
func GetGroups(ldapc Connection) []Group {

	if len(groups) == 0 {
		fillRoles(ldapc)
		fillGroups(ldapc)
	}

	return groups
}

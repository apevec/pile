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
	UID       string
	Name      string
	Role      string
	Squad     string
	Component string
	External  map[string]string
	IRC       string
	Location  string
}

// Group defines the DFG model.
type Group struct {
	Name       string
	Desc       string
	Members    []string
	Backlog    string
	Mission    string
	Links      map[string]string
	PMs        []Member
	Stewards   []Member
	UAs        []Member
	TCs        []Member
	SquadLeads []Member
	Squads     map[string]string
	SquadsSz   int
}

var groups []Group
var people = map[string]*Member{}

// Connection is an interface for making queries.
type Connection interface {
	Search(*ldap.SearchRequest) (*ldap.SearchResult, error)
}

func removeDuplicates(xs *[]string) {
	found := make(map[string]bool)
	j := 0
	for i, x := range *xs {
		if !found[x] {
			found[x] = true
			(*xs)[j] = (*xs)[i]
			j++
		}
	}
	*xs = (*xs)[:j]
}

// decodeNote - returns kv
func decodeNote(note string) map[string]string {
	result := make(map[string]string)

	re, _ := regexp.Compile(`pile:(\w*=[a-zA-z0-9:/.@-]+)`)
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
		members := rolesPeople.GetAttributeValues("memberUid")
		for _, person := range members {
			if _, ok := people[person]; !ok {
				people[person] = &Member{}
			}
			people[person] = &Member{}
			people[person].UID = person

			// TODO: remove
			if person == "aarapov" {
				people[person].Role = "Steward"
				continue
			}

			switch role {
			case "rhos-pm":
				people[person].Role = "Product Manager"
			case "rhos-ua":
				people[person].Role = "User Advocate"
			case "rhos-tc":
				people[person].Role = "Team Catalyst"
			case "rhos-stewards-em":
				fallthrough
			case "rhos-stewards-qe":
				fallthrough
			case "rhos-stewards":
				people[person].Role = "Steward"
			case "rhos-squad-lead":
				people[person].Role = "Squad Lead"
			default:
				people[person].Role = "Engineer"
			}
		}

		fillMembers(ldapc, members)
	}
}

func fillMembers(ldapc Connection, members []string) {

	filter := "(&(objectClass=rhatPerson)(|"
	for _, member := range members {
		// "(&(objectClass=rhatPerson)(|(uid=user1)(uid=user2)(uid=user3)))"

		// don't do it multiple times
		if _, ok := people[member]; ok {
			if people[member].Name != "" {
				continue
			}
		}

		filter = filter + fmt.Sprintf("(uid=%s)", member)
	}
	filter = filter + "))"

	sMembersRequest := ldap.NewSearchRequest(
		"ou=users,dc=redhat,dc=com",
		ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false,
		filter, // The filter to apply
		[]string{"uid", "cn", "co", "rhatBio", "rhatNickName"},
		nil,
	)

	ldapMembers, _ := ldapc.Search(sMembersRequest)
	// TODO: check for err
	for _, ldapMember := range ldapMembers.Entries {
		uid := ldapMember.GetAttributeValue("uid")
		if _, ok := people[uid]; !ok {
			people[uid] = &Member{}
		}

		people[uid].Name = ldapMember.GetAttributeValue("cn")

		kv := decodeNote(ldapMember.GetAttributeValue("rhatBio"))
		people[uid].External = make(map[string]string)
		for k, v := range kv {
			switch k {
			case "components":
				people[uid].Component = v
			case "gtalk":
				people[uid].External[strings.Title(k)] = v
			}
		}

		people[uid].IRC = ldapMember.GetAttributeValue("rhatNickName")
		people[uid].Location = ldapMember.GetAttributeValue("co")
		if people[uid].Role == "" {
			people[uid].Role = "Engineer"
		}
	}
}

func fillGroups(ldapc Connection) {
	sGroupRequest := ldap.NewSearchRequest(
		"ou=adhoc,ou=managedGroups,dc=redhat,dc=com",
		ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectClass=rhatGroup)(&(cn=rhos-dfg-*)(!(cn=*squad*))(!(cn=rhos-*lt*))))",
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
			group.Links = make(map[string]string)
			for k, v := range kv {
				switch k {
				case "backlog":
					group.Backlog = v
				case "mission":
					group.Mission = v
				default:
					group.Links[strings.Title(k)] = v
				}
			}
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
				squadMembers := ldapSquad.GetAttributeValues("memberUid")
				fillMembers(ldapc, squadMembers)
				for i, squadMember := range squadMembers {
					// TODO: remove
					if squadMember == "aarapov" {
						squadMembers = append(squadMembers[:i], squadMembers[i+1:]...)
						continue
					}
					people[squadMember].Squad = ldapSquad.GetAttributeValue("description")
				}

				group.Members = append(group.Members, squadMembers...)
				removeDuplicates(&group.Members)
			}
		}

		group.Name = name
		group.Desc = ldapGroup.GetAttributeValue("description")

		groupMembers := ldapGroup.GetAttributeValues("memberUid")
		for i, groupMember := range groupMembers {
			if (name != "rhos-dfg-cloud-applications") && (name != "rhos-dfg-portfolio-integration") {
				if groupMember == "aarapov" {
					groupMembers = append(groupMembers[:i], groupMembers[i+1:]...)
					break
				}
			}
		}

		for _, groupMember := range groupMembers {
			if _, ok := people[groupMember]; ok {
				switch people[groupMember].Role {
				case "Product Manager":
					group.PMs = append(group.PMs, *people[groupMember])
				case "Steward":
					group.Stewards = append(group.Stewards, *people[groupMember])
				case "User Advocate":
					group.UAs = append(group.UAs, *people[groupMember])
				case "Team Catalyst":
					group.TCs = append(group.TCs, *people[groupMember])
				case "Squad Lead":
					group.SquadLeads = append(group.SquadLeads, *people[groupMember])
				}
			}
		}
		group.Members = append(group.Members, groupMembers...)
		groups = append(groups, group)
	}
}

// GetMembers - tbd
func GetMembers(ldapc Connection, group string) map[string]*Member {
	var groupMembers = map[string]*Member{}

	for _, grp := range groups {
		if grp.Name == group {
			fillMembers(ldapc, grp.Members)
			for _, member := range grp.Members {
				groupMembers[member] = people[member]
			}
			return groupMembers
		}
	}

	return nil
}

// GetGroups - tbd
func GetGroups(ldapc Connection) []Group {

	if len(groups) == 0 {
		fillRoles(ldapc)
		fillGroups(ldapc)
	}

	return groups
}

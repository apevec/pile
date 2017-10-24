package ldapxrest

import (
	ldap "gopkg.in/ldap.v2"
)

type Connection interface {
	GetAllGroupsTiny() ([]*ldap.Entry, error)
	GetAllSquadsTiny(group string) ([]*ldap.Entry, error)
	GetGroupMembers(group string) (*ldap.Entry, error)
	GetSquadMembers(group string, squad string) (*ldap.Entry, error)
}

func GetGroupMembers(ldapc Connection, group string) (map[string]string, int, error) {
	var members = make(map[string]string)

	ldapGroupMembers, err := ldapc.GetGroupMembers(group)
	if err != nil {
		return members, 0, err
	}
	groupMembers := ldapGroupMembers.GetAttributeValues("memberUid")

	squads, err := GetSquads(ldapc, group)
	if err != nil {
		return members, 0, err
	}
	for squad := range squads {
		ldapSquadMembers, _ := ldapc.GetSquadMembers(group, squad)
		// TODO: handle error gracefully

		squadMembers := ldapSquadMembers.GetAttributeValues("memberUid")
		groupMembers = append(groupMembers, squadMembers...)
	}

	removeDuplicates(&groupMembers)
	for _, groupMember := range groupMembers {
		members[groupMember] = "tbd: name"
	}

	return members, len(squads), err
}

func GetGroupSize(ldapc Connection, group string) (map[string]int, error) {
	var size = make(map[string]int)

	groupMembers, squads, err := GetGroupMembers(ldapc, group)
	if err != nil {
		return size, err
	}

	size["people"] = len(groupMembers)
	size["squads"] = squads

	return size, err
}

func GetSquads(ldapc Connection, group string) (map[string]string, error) {
	var squads = make(map[string]string)

	ldapSquads, err := ldapc.GetAllSquadsTiny(group)
	if err != nil {
		return nil, err
	}

	for _, ldapSquad := range ldapSquads {
		squadName := ldapSquad.GetAttributeValue("cn")
		squadDesc := ldapSquad.GetAttributeValue("description")

		squads[squadName] = squadDesc
	}

	return squads, err
}

func GetGroups(ldapc Connection) (map[string]string, error) {
	var groups = make(map[string]string)

	ldapGroups, err := ldapc.GetAllGroupsTiny()
	if err != nil {
		return groups, err
	}

	for _, ldapGroup := range ldapGroups {
		groupName := ldapGroup.GetAttributeValue("cn")
		groupDesc := ldapGroup.GetAttributeValue("description")

		groups[groupName] = groupDesc
	}

	return groups, err
}

// helpers
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

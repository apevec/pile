package ldapxrest

import (
	ldap "gopkg.in/ldap.v2"
)

type role struct {
	Name    string
	Members []string
}

type Connection interface {
	GetAllRoles() ([]*ldap.Entry, error)
	GetAllGroupsTiny() ([]*ldap.Entry, error)
	GetAllSquadsTiny(group string) ([]*ldap.Entry, error)
	GetGroupMembers(group string) (*ldap.Entry, error)
	GetSquadMembers(group string, squad string) (*ldap.Entry, error)
	GetPeopleTiny(ids []string) ([]*ldap.Entry, error)
}

func GetGroupHead(ldapc Connection, group string) (map[string][]map[string]string, error) {
	var head = make(map[string][]map[string]string) // head["role"][...]["ID"] = uid

	roles, err := GetRoles(ldapc)
	if err != nil {
		return head, err
	}

	var mapPeopleRole = make(map[string]string)
	var mapPeopleName = make(map[string]string)
	for _, role := range roles {
		people, _ := GetPeople(ldapc, role.Members)
		// TODO: handle error gracefully

		for uid, name := range people {
			mapPeopleRole[uid] = role.Name
			mapPeopleName[uid] = name
		}

	}

	groupMembers, err := GetGroupMembers(ldapc, group)
	if err != nil {
		return head, err
	}
	for _, uid := range groupMembers {
		if _, ok := mapPeopleRole[uid]; !ok {
			continue // skip members who doesn't belong to any role
		}

		role := mapPeopleRole[uid]
		name := mapPeopleName[uid]
		info := map[string]string{"ID": uid, "Name": name}

		head[role] = append(head[role], info)
	}

	return head, err
}

func GetPeople(ldapc Connection, uids []string) (map[string]string, error) {
	var people = make(map[string]string)

	ldapPeople, err := ldapc.GetPeopleTiny(uids)
	if err != nil {
		return people, err
	}
	for _, ldapMan := range ldapPeople {
		uid := ldapMan.GetAttributeValue("uid")
		fullname := ldapMan.GetAttributeValue("cn")

		people[uid] = fullname
	}

	return people, err
}

func GetRoles(ldapc Connection) (map[string]*role, error) {
	var roles = map[string]*role{}

	ldapRoles, err := ldapc.GetAllRoles()
	if err != nil {
		return roles, err
	}

	for _, ldapRole := range ldapRoles {
		roleID := ldapRole.GetAttributeValue("cn")
		roleName := ldapRole.GetAttributeValue("description")
		roleMembers := ldapRole.GetAttributeValues("memberUid")

		// TODO: find a better way for exclusions
		if roleID != "rhos-steward" {
			removeMe(&roleMembers)
		}
		roles[roleID] = &role{
			Name:    roleName,
			Members: roleMembers,
		}
	}

	return roles, err
}

func GetGroupMembers(ldapc Connection, group string) ([]string, error) {
	var members []string

	ldapGroupMembers, err := ldapc.GetGroupMembers(group)
	if err != nil {
		return members, err
	}
	groupMembers := ldapGroupMembers.GetAttributeValues("memberUid")

	squads, err := GetSquads(ldapc, group)
	if err != nil {
		return members, err
	}
	for squad := range squads {
		ldapSquadMembers, _ := ldapc.GetSquadMembers(group, squad)
		// TODO: handle error gracefully

		squadMembers := ldapSquadMembers.GetAttributeValues("memberUid")
		groupMembers = append(groupMembers, squadMembers...)
	}

	removeDuplicates(&groupMembers)

	// TODO: find a better way for exclusion
	if (group != "rhos-dfg-cloud-applications") && (group != "rhos-dfg-portfolio-integration") {
		removeMe(&groupMembers)
	}
	members = groupMembers

	return members, err
}

func GetGroupSize(ldapc Connection, group string) (map[string]int, error) {
	var size = make(map[string]int)

	groupMembers, err := GetGroupMembers(ldapc, group)
	if err != nil {
		return size, err
	}
	squads, err := GetSquads(ldapc, group)
	if err != nil {
		return size, err
	}

	size["people"] = len(groupMembers)
	size["squads"] = len(squads)

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

func Ping(ldapc Connection) (map[string]string, error) {
	ldapMe, err := ldapc.GetPeopleTiny([]string{"aarapov"})
	if err != nil {
		return nil, err
	}

	pong := map[string]string{
		"uid":  ldapMe[0].GetAttributeValue("uid"),
		"name": ldapMe[0].GetAttributeValue("cn"),
	}

	return pong, err
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

func removeMe(xs *[]string) {
	// TODO: temporary, remove aarapov
	for i, me := range *xs {
		if me == "aarapov" {
			(*xs) = append((*xs)[:i], (*xs)[i+1:]...)
			break
		}
	}
}

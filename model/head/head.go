// Package head
package head

import ldap "gopkg.in/ldap.v2"
import "log"

// Item defines the model.
type Head struct {
	GroupName string
	Role      string
	Name      string
	UID       string
}

type Role struct {
	Name string
	Desc string
}

var (
	head          []Head
	mapMemberRole = map[string]*Role{}
	mapMemberName = make(map[string]string)
)

// Connection is an interface for making queries.
type Connection interface {
	GetAllGroups() ([]*ldap.Entry, error)
	GetAllRoles() ([]*ldap.Entry, error)
	GetMembersTiny(ids []string) ([]*ldap.Entry, error)
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

func removeMe(xs *[]string) {
	// TODO: temporary, remove aarapov
	for i, me := range *xs {
		if me == "aarapov" {
			(*xs) = append((*xs)[:i], (*xs)[i+1:]...)
			break
		}
	}
}

func GetHead(ldapc Connection) ([]Head, error) {
	var allmembers []string

	ldapRoles, err := ldapc.GetAllRoles()
	if err != nil {
		log.Println(err)
		return head, err
	}
	for _, ldapRole := range ldapRoles {
		id := ldapRole.GetAttributeValue("cn")
		desc := ldapRole.GetAttributeValue("description")
		members := ldapRole.GetAttributeValues("memberUid")

		// TODO: removeme
		if id != "rhos-steward" {
			removeMe(&members)
		}

		for _, member := range members {
			mapMemberRole[member] = &Role{id, desc}
		}

		allmembers = append(allmembers, members...)
		removeDuplicates(&allmembers)
	}

	ldapMembers, err := ldapc.GetMembersTiny(allmembers)
	if err != nil {
		log.Println(err)
		return head, err
	}
	for _, ldapMember := range ldapMembers {
		id := ldapMember.GetAttributeValue("uid")
		name := ldapMember.GetAttributeValue("cn")
		mapMemberName[id] = name
	}

	ldapGroups, err := ldapc.GetAllGroups()
	if err != nil {
		log.Println(err)
		return head, err
	}
	for _, ldapGroup := range ldapGroups {
		id := ldapGroup.GetAttributeValue("cn")
		groupName := ldapGroup.GetAttributeValue("description")
		groupMembers := ldapGroup.GetAttributeValues("memberUid")

		// TODO: removeme
		if (id != "rhos-dfg-cloud-applications") && (id != "rhos-dfg-portfolio-integration") {
			removeMe(&groupMembers)
		}

		for _, groupMember := range groupMembers {
			if _, ok := mapMemberRole[groupMember]; !ok {
				continue // skip members who doesn't belong to any role
			}

			headi := Head{
				GroupName: groupName,
				Role:      mapMemberRole[groupMember].Desc,
				Name:      mapMemberName[groupMember],
				UID:       groupMember,
			}

			head = append(head, headi)
		}
	}

	return head, err
}

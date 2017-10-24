package ldapxrest

import ldap "gopkg.in/ldap.v2"

type Connection interface {
	GetAllGroupsTiny() ([]*ldap.Entry, error)
}

func GetGroups(ldapc Connection) (map[string]string, error) {
	var groups = make(map[string]string)

	ldapGroups, err := ldapc.GetAllGroupsTiny()
	if err != nil {
		return nil, err
	}

	for _, ldapGroup := range ldapGroups {
		groupName := ldapGroup.GetAttributeValue("cn")
		groupDesc := ldapGroup.GetAttributeValue("description")

		groups[groupName] = groupDesc
	}

	return groups, err
}

// Package ldapx
package ldapx

import (
	"fmt"
	"log"

	ldap "gopkg.in/ldap.v2"
)

var (
	basednGroups  = "ou=adhoc,ou=managedGroups,dc=redhat,dc=com"
	basednMembers = "ou=users,dc=redhat,dc=com"

	ldapAttrGroup = []string{
		"cn",             // group id
		"description",    // group description
		"memberUid",      // []members
		"rhatGroupNotes", // group notes
	}
	ldapAttrMember = []string{
		"cn", // idk
	}
)

// Info holds the config.
type Info struct {
	Hostname string
	Port     int
}

type Conn struct {
	*ldap.Conn
}

func (c Info) Dial() (*Conn, error) {
	//return ldapx.Dial("tcp", fmt.Sprintf("%s:%d", c.Hostname, c.Port))
	parentConn, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", c.Hostname, c.Port))
	return &Conn{parentConn}, err
}

// rhos dfg in ldap should keep the following schema:
// - rhos-dfg-[name_of_group]
// whereas squad[s] that belong to the group:
// - rhos-dfg-[name_of_group]-squad-[name_of_squad]
func (c *Conn) query(basedn string, ldapAttributes []string, filter string) ([]*ldap.Entry, error) {

	sGroupRequest := ldap.NewSearchRequest(
		basedn, ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false,
		filter, ldapAttributes, nil,
	)
	ldapGroups, err := c.Search(sGroupRequest)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return ldapGroups.Entries, err
}

func (c *Conn) GetGroups(groups ...string) ([]*ldap.Entry, error) {
	var filter string

	if len(groups) == 0 {
		// "(&(objectClass=rhatGroup)(&(cn=rhos-dfg-*)(!(cn=*squad*))))"
		filter = "(&(objectClass=rhatGroup)(&(cn=rhos-dfg-*)(!(cn=*squad*))))"
	} else {
		// "(&(objectClass=rhatGroup)(|(cn=group1)(uid=group2)(uid=group3)))"
		filter = "(&(objectClass=rhatGroup)(&"
		for _, group := range groups {
			filter = filter + fmt.Sprintf("(cn=%s)", group)
		}
		filter = filter + "))"
	}

	return c.query(basednGroups, ldapAttrGroup, filter)
}

func (c *Conn) GetGroup(group string) (*ldap.Entry, error) {
	ldapGroups, err := c.GetGroups(group)
	return ldapGroups[0], err
}

func (c *Conn) GetAllGroups() ([]*ldap.Entry, error) {
	return c.GetGroups()
}

func (c *Conn) GetAllSquads(group string) ([]*ldap.Entry, error) {
	var filter string

	// "(&(objectClass=rhatGroup)(cn=%s-squad-*))"
	filter = fmt.Sprintf("(&(objectClass=rhatGroup)(cn=%s-squad-*))", group)

	return c.query(basednGroups, ldapAttrGroup, filter)
}

//func GetSquads(ldapc *ldap.Conn)

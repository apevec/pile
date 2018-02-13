// Package ldapx
package ldapx

import (
	"fmt"
	"log"

	ldap "gopkg.in/ldap.v2"
)

var (
	// TODO: get some/all of those to configuration
	basednGroups  = "ou=adhoc,ou=managedGroups,dc=redhat,dc=com"
	basednMembers = "ou=users,dc=redhat,dc=com"

	ldapAttrGroupTiny = []string{
		"cn",          // group id
		"description", // description
	}
	ldapAttrGroupMembers = []string{
		"uniqueMember", // []members
		"owner",        // []members owners
	}
	ldapAttrGroupLinks = []string{
		"rhatGroupNotes", // notes
	}
	ldapAttrGroup = []string{
		"cn",           // group id
		"description",  // description
		"uniqueMember", // []members
	}
	ldapAttrRoles = []string{
		"cn",           // roles id
		"description",  // description
		"uniqueMember", // []members
	}
	ldapAttrMemberTiny = []string{
		"uid",            // uid
		"cn",             // fullname
		"rhatCostCenter", // cost center
	}
	ldapAttrMemberFull = []string{
		"uid",                // uid
		"cn",                 // fullname
		"co",                 // country
		"rhatBio",            // notes/bio
		"rhatNickName",       // irc nick
		"rhatCostCenter",     // cost center
		"rhatLocation",       // location
		"rhatOfficeLocation", // describes REMOTE
		"mobile",             // mobile phone
	}
	ldapAttrMemberLocation = []string{
		"uid",                // uid
		"registeredAddress",  // lat/lng
		"rhatLocation",       // location
		"rhatOfficeLocation", // describes REMOTE
	}

	// ldapRolesMap - keys are the role groups in ldap
	ldapRolesMap = map[string]string{
		"rhos-pm":         "Product Management",
		"rhos-steward":    "Steward",
		"rhos-ua":         "User Advocate",
		"rhos-tc":         "Team Catalyst",
		"rhos-squad-lead": "Squad Lead",
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
	// TODO: every minute calls of uri/ping is keeping pile connected to ldap
	// in openshift. Disconects were not observed when run on local machine.
	// - It may need ReDial() function...

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

func (c *Conn) getGroups(ldapAttributes []string, groups ...string) ([]*ldap.Entry, error) {
	var filter string

	// "(&(objectClass=rhatRoverGroup)(&(cn=rhos-dfg-*)(!(cn=*squad*))))"
	filter = "(&(objectClass=rhatRoverGroup)(&(cn=rhos-dfg-*)(!(cn=*squad*))))"
	if len(groups) > 0 && groups[0] != "" {
		filter = "(&(objectClass=rhatRoverGroup)(&"
		for _, group := range groups {
			filter += fmt.Sprintf("(cn=%s)", group)
		}
		filter += "))"
	}

	return c.query(basednGroups, ldapAttributes, filter)
}

func (c *Conn) GetGroupsTiny(groups ...string) ([]*ldap.Entry, error) {
	return c.getGroups(ldapAttrGroupTiny, groups...)
}

func (c *Conn) GetAllGroups() ([]*ldap.Entry, error) {
	return c.getGroups(ldapAttrGroup)
}

func (c *Conn) GetGroupMembers(group string) (*ldap.Entry, error) {
	ldapGroups, err := c.getGroups(ldapAttrGroupMembers, group)
	return ldapGroups[0], err
}

func (c *Conn) GetGroupLinks(group string) (*ldap.Entry, error) {
	ldapLinks, err := c.getGroups(ldapAttrGroupLinks, group)
	return ldapLinks[0], err
}

func (c *Conn) getSquads(ldapAttributes []string, group string, squads ...string) ([]*ldap.Entry, error) {
	var filter string

	// "(&(objectClass=rhatRoverGroup)(cn=rhos-dfg-%group%-squad-*))"
	filter = fmt.Sprintf("(&(objectClass=rhatRoverGroup)(cn=%s-squad-*))", group)
	if len(squads) > 0 {
		filter = "(&(objectClass=rhatRoverGroup)(&"
		for _, squad := range squads {
			filter += fmt.Sprintf("(cn=%s)", squad)
		}
		filter += "))"
	}

	return c.query(basednGroups, ldapAttributes, filter)
}

func (c *Conn) GetAllSquads(group string) ([]*ldap.Entry, error) {
	return c.getSquads(ldapAttrGroup, group)
}

func (c *Conn) GetAllSquadsTiny(group string) ([]*ldap.Entry, error) {
	return c.getSquads(ldapAttrGroupTiny, group)
}

func (c *Conn) GetSquadMembers(group string, squad string) (*ldap.Entry, error) {
	ldapSquads, err := c.getSquads(ldapAttrGroupMembers, group, squad)
	return ldapSquads[0], err
}

func (c *Conn) getRoles(roles ...string) ([]*ldap.Entry, error) {
	var filter string

	// "(&(objectClass=rhatGroup)(|(cn=rhos-role1)(cn=rhos-role2)))"
	filter = "(&(objectClass=rhatGroup)(|"
	for ldapRoleGroup := range ldapRolesMap {
		filter = filter + fmt.Sprintf("(cn=%s)", ldapRoleGroup)
	}
	filter = filter + "))"

	return c.query(basednGroups, ldapAttrRoles, filter)
}

func (c *Conn) GetAllRoles() ([]*ldap.Entry, error) {
	return c.getRoles()
}

func (c *Conn) getPeople(ldapAttributes []string, ids []string) ([]*ldap.Entry, error) {
	var filter string

	// "(&(objectClass=rhatPerson)(|(uid=user1)(uid=user2)(uid=user3)))"
	filter = "(&(objectClass=rhatPerson)(|"
	for _, id := range ids {
		filter = filter + fmt.Sprintf("(uid=%s)", id)
	}
	filter = filter + "))"

	return c.query(basednMembers, ldapAttributes, filter)
}

func (c *Conn) GetPeopleTiny(ids []string) ([]*ldap.Entry, error) {
	return c.getPeople(ldapAttrMemberTiny, ids)
}

func (c *Conn) GetPeopleFull(ids []string) ([]*ldap.Entry, error) {
	return c.getPeople(ldapAttrMemberFull, ids)
}

func (c *Conn) GetPeopleLocationData(ids ...string) ([]*ldap.Entry, error) {
	return c.getPeople(ldapAttrMemberLocation, ids)
}

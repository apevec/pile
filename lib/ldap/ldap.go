// Package ldap
package ldap

import (
	"fmt"

	"gopkg.in/ldap.v2"
)

// Info holds the config.
type Info struct {
	Hostname string
	Port     int
}

func (c Info) Dial() (*ldap.Conn, error) {
	return ldap.Dial("tcp", fmt.Sprintf("%s:%d", c.Hostname, c.Port))
}

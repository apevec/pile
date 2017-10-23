// Package roster - tbd
package roster

import (
	"context"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"googlemaps.github.io/maps"

	ldap "gopkg.in/ldap.v2"
)

// Member - tbd
type Member struct {
	Name     string
	Role     string
	Squad    string
	Data     map[string]string
	IRC      string
	CC       string
	Country  string
	UTC      string
	Timezone string
	Remote   bool
}

type Group struct {
	Name   string
	Links  map[string]string
	Head   map[string][]map[string]string
	Squads int
	Size   int

	members []string
}

type Role struct {
	Name string
	Desc string
}

var (
	groups = map[string]*Group{}

	mapMemberRole  = map[string]*Role{}
	mapMemberName  = make(map[string]string)
	mapMemberSquad = make(map[string]string)
)

// Connection is an interface for making queries.
type Connection interface {
	GetAllGroups() ([]*ldap.Entry, error)
	GetAllSquads(group string) ([]*ldap.Entry, error)
	GetAllRoles() ([]*ldap.Entry, error)
	GetMembersTiny(ids []string) ([]*ldap.Entry, error)
	GetMembersFull(ids []string) ([]*ldap.Entry, error)
}

func decodeNote(note string) map[string]string {
	result := make(map[string]string)

	re, _ := regexp.Compile(`pile:(\w*=[a-zA-z0-9:/.@-]+)`)
	// TODO: take care of error here
	pile := re.FindAllStringSubmatch(note, -1)
	// TODO: code below is fragile, very fragile
	for i := range pile {
		kv := strings.Split(pile[i][1], "=")
		result[strings.Title(kv[0])] = kv[1]
	}

	return result
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

func getTimeZone(latlng string, location string, remote bool) (string, string) {
	utc := "n/a"
	timezone := "undefined"

	// TODO: take this out to configuration
	gapi := os.Getenv("GAPI")
	if gapi == "" {
		log.Println("GAPI environment variable is not set. Can't find timezone!")
		return utc, timezone
	}
	mapsc, err := maps.NewClient(maps.WithAPIKey(gapi))
	if err != nil {
		log.Println(err)
		return utc, timezone
	}

	var lat float64
	var lng float64
	if remote == true {
		locationTrim1 := strings.Replace(location, "RH -", "", 1)
		locationTrim2 := strings.Replace(locationTrim1, "Remote ", "", 1)

		// fallback, if no latitude and longitude are known
		r := &maps.GeocodingRequest{
			Address: locationTrim2,
		}

		loc, err := mapsc.Geocode(context.Background(), r)
		if err != nil {
			log.Println(err)
			return utc, timezone
		}

		lat = loc[0].Geometry.Location.Lat
		lng = loc[0].Geometry.Location.Lng
	} else {
		lat, _ = strconv.ParseFloat(strings.Split(latlng, ",")[0], 64)
		lng, _ = strconv.ParseFloat(strings.Split(latlng, ",")[1], 64)
	}

	r := &maps.TimezoneRequest{
		Location: &maps.LatLng{
			Lat: lat,
			Lng: lng,
		},
		Timestamp: time.Now().UTC(),
	}

	tz, err := mapsc.Timezone(context.Background(), r)
	if err != nil {
		log.Println(err)
		return utc, timezone
	}

	utcOffset := (tz.RawOffset + tz.DstOffset) / 3600
	utc = strconv.Itoa(utcOffset)
	if utcOffset >= 0 {
		utc = "+" + utc
	}
	timezone = tz.TimeZoneName

	return utc, timezone
}

func GetGroups(ldapc Connection) (map[string]*Group, error) {
	var allmembers []string

	ldapRoles, err := ldapc.GetAllRoles()
	if err != nil {
		log.Println(err)
		return nil, err
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
		return nil, err
	}
	for _, ldapMember := range ldapMembers {
		id := ldapMember.GetAttributeValue("uid")
		name := ldapMember.GetAttributeValue("cn")
		mapMemberName[id] = name
	}

	ldapGroups, err := ldapc.GetAllGroups()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	for _, ldapGroup := range ldapGroups {
		id := ldapGroup.GetAttributeValue("cn")
		desc := ldapGroup.GetAttributeValue("description")
		members := ldapGroup.GetAttributeValues("memberUid")
		links := decodeNote(ldapGroup.GetAttributeValue("rhatGroupNotes"))
		head := make(map[string][]map[string]string) // head["role"][...]["ID"] = uid
		squads := 0

		// TODO: removeme
		if (id != "rhos-dfg-cloud-applications") && (id != "rhos-dfg-portfolio-integration") {
			removeMe(&members)
		}

		for _, member := range members {
			if _, ok := mapMemberRole[member]; !ok {
				continue // skip members who doesn't belong to any role
			}

			role := mapMemberRole[member].Desc
			name := mapMemberName[member]
			info := map[string]string{"ID": member, "Name": name}

			head[role] = append(head[role], info)
		}

		// Squads
		ldapSquads, err := ldapc.GetAllSquads(id)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		for _, ldapSquad := range ldapSquads {
			squads++
			squad := ldapSquad.GetAttributeValue("description")
			membersSquad := ldapSquad.GetAttributeValues("memberUid")

			// TODO: removeme
			removeMe(&membersSquad)

			for _, member := range membersSquad {
				mapMemberSquad[member] = squad
			}

			members = append(members, membersSquad...)
			removeDuplicates(&members)
		}

		groups[id] = &Group{
			Name:    desc,
			Links:   links,
			Head:    head,
			Squads:  squads,
			Size:    len(members),
			members: members,
		}
	}

	return groups, err
}

func GetMembers(ldapc Connection, group string) (map[string]*Member, error) {
	var members = map[string]*Member{}

	ldapMembers, err := ldapc.GetMembersFull(groups[group].members)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	for _, ldapMember := range ldapMembers {
		uid := ldapMember.GetAttributeValue("uid")
		name := ldapMember.GetAttributeValue("cn")
		data := decodeNote(ldapMember.GetAttributeValue("rhatBio"))
		ircnick := ldapMember.GetAttributeValue("rhatNickName")
		cc := ldapMember.GetAttributeValue("rhatCostCenter")
		country := ldapMember.GetAttributeValue("co")

		remote := false
		if ldapMember.GetAttributeValue("rhatOfficeLocation") == "REMOTE" {
			remote = true
		}

		latlng := ldapMember.GetAttributeValue("registeredAddress")
		location := ldapMember.GetAttributeValue("rhatLocation")
		utc, timezone := getTimeZone(latlng, location, remote)

		role := "Engineer"
		if _, ok := mapMemberRole[uid]; ok {
			role = mapMemberRole[uid].Desc
		}

		squad := ""
		if _, ok := mapMemberSquad[uid]; ok {
			squad = mapMemberSquad[uid]
		}

		members[uid] = &Member{
			Name:     name,
			Role:     role,
			Data:     data,
			Squad:    squad,
			IRC:      ircnick,
			CC:       cc,
			Country:  country,
			UTC:      utc,
			Timezone: timezone,
			Remote:   remote,
		}
	}

	return members, err
}

func Ping(ldapc Connection) (map[string]string, error) {
	ldapMe, err := ldapc.GetMembersTiny([]string{"aarapov"})
	if err != nil {
		log.Println(err)
		return nil, err
	}

	pong := map[string]string{
		"uid":  ldapMe[0].GetAttributeValue("uid"),
		"name": ldapMe[0].GetAttributeValue("cn"),
	}

	return pong, err
}

package ldapxrest

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

type role struct {
	Name    string
	Members []string
}

type member struct {
	Name    string
	Role    string
	Squad   string
	Data    map[string]string
	IRC     string
	Country string
	CC      string
	Remote  bool
}

type Connection interface {
	GetAllRoles() ([]*ldap.Entry, error)
	GetAllSquadsTiny(group string) ([]*ldap.Entry, error)
	GetGroupMembers(group string) (*ldap.Entry, error)
	GetSquadMembers(group string, squad string) (*ldap.Entry, error)
	GetPeopleTiny(ids []string) ([]*ldap.Entry, error)
	GetPeopleFull(ids []string) ([]*ldap.Entry, error)
	GetPeopleLocationData(ids ...string) ([]*ldap.Entry, error)
	GetGroupLinks(group string) (*ldap.Entry, error)
	GetGroupsTiny(groups ...string) ([]*ldap.Entry, error)
}

func GetTimezoneInfo(ldapc Connection, uid string) (map[string]string, error) {
	var tzinfo = make(map[string]string)

	ldapLocationData, err := ldapc.GetPeopleLocationData(uid)
	if err != nil {
		return tzinfo, err
	}

	ldapLocation := ldapLocationData[0] // safe: we have alays one item here

	remote := false
	if ldapLocation.GetAttributeValue("rhatOfficeLocation") == "REMOTE" {
		remote = true
	}

	latlng := ldapLocation.GetAttributeValue("registeredAddress")
	location := ldapLocation.GetAttributeValue("rhatLocation")
	utc, timezone, lat, lng, err := getTimeZone(latlng, location, remote)
	if err != nil {
		return nil, err
	}

	tzinfo["utcOffset"] = utc
	tzinfo["tzName"] = timezone
	tzinfo["remote"] = strconv.FormatBool(remote)
	tzinfo["latlng"] = strconv.FormatFloat(lat, 'f', -1, 64) + "," + strconv.FormatFloat(lng, 'f', -1, 64)

	return tzinfo, err
}

func GetGroupMembers(ldapc Connection, group string) (map[string]*member, error) {
	var members = map[string]*member{}

	uids, err := GetGroupMembersSlice(ldapc, group)
	if err != nil {
		return members, err
	}

	roles, err := GetRoles(ldapc)
	if err != nil {
		return members, err
	}

	var mapPeopleRole = make(map[string]string)
	for _, role := range roles {
		for _, uid := range role.Members {
			mapPeopleRole[uid] = role.Name
		}
	}

	ldapPeople, err := ldapc.GetPeopleFull(uids)
	if err != nil {
		return members, err
	}

	var mapPeopleSquad = make(map[string]string)
	squads, err := GetSquads(ldapc, group)
	if err != nil {
		return members, err
	}
	for squad, squadName := range squads {
		squadMembers, _ := GetGroupMembersSlice(ldapc, squad)
		for _, squadMember := range squadMembers {

			if mapPeopleRole[squadMember] == "Steward" {
				// We don't want Steward to be a member of any squad.
				// So we skip squad assignment here
				continue
			}

			mapPeopleSquad[squadMember] = squadName
		}
	}

	for _, man := range ldapPeople {
		uid := man.GetAttributeValue("uid")
		name := man.GetAttributeValue("cn")
		ircnick := man.GetAttributeValue("rhatNickName")
		data := decodeNote(man.GetAttributeValue("rhatBio"))
		co := man.GetAttributeValue("co")
		cc := man.GetAttributeValue("rhatCostCenter")

		remote := false
		if man.GetAttributeValue("rhatOfficeLocation") == "REMOTE" {
			remote = true
		}

		role := "Engineer"
		if _, ok := mapPeopleRole[uid]; ok {
			role = mapPeopleRole[uid]
		}

		squad := ""
		if _, ok := mapPeopleSquad[uid]; ok {
			squad = mapPeopleSquad[uid]
		}

		members[uid] = &member{
			Name:    name,
			Role:    role,
			Squad:   squad,
			Data:    data,
			IRC:     ircnick,
			Country: co,
			CC:      cc,
			Remote:  remote,
		}

	}

	return members, err
}

func GetGroupLinks(ldapc Connection, group string) (map[string]string, error) {
	var links = make(map[string]string)

	ldapLinks, err := ldapc.GetGroupLinks(group)
	if err != nil {
		return links, err
	}
	links = decodeNote(ldapLinks.GetAttributeValue("rhatGroupNotes"))

	return links, err
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

	groupMembers, err := GetGroupMembersSlice(ldapc, group)
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

func GetGroupMembersSlice(ldapc Connection, group string) ([]string, error) {
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

	groupMembers, err := GetGroupMembersSlice(ldapc, group)
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

func GetGroups(ldapc Connection, groups ...string) (map[string]string, error) {
	var res = make(map[string]string)

	ldapGroups, err := ldapc.GetGroupsTiny(groups...)
	if err != nil {
		return res, err
	}

	for _, ldapGroup := range ldapGroups {
		groupName := ldapGroup.GetAttributeValue("cn")
		groupDesc := ldapGroup.GetAttributeValue("description")

		res[groupName] = groupDesc
	}

	return res, err
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

func getTimeZone(latlng string, location string, remote bool) (string, string, float64, float64, error) {
	utc := ""
	timezone := ""

	// TODO: take this out to configuration
	gapi := os.Getenv("GAPI")
	if gapi == "" {
		log.Println("GAPI environment variable is not set. Can't find timezone!")
		// return utc, timezone
	}
	mapsc, err := maps.NewClient(maps.WithAPIKey(gapi))
	if err != nil {
		log.Println(err)
		return utc, timezone, 0, 0, err
	}

	var lat float64
	var lng float64
	if remote == true {
		// Remotes doesn't have Lat/Lng set in LDAP, thus we have to guess it
		// based on rhatLocation field

		// TODO: Put some nice regexp here?
		locationTrim1 := strings.Replace(location, "RH -", "", 1)
		locationTrim2 := strings.Replace(locationTrim1, "Remote ", "", 1)
		locationTrim3 := strings.Replace(locationTrim2, "US", "USA", 1)

		r := &maps.GeocodingRequest{
			Address: locationTrim3,
		}
		loc, err := mapsc.Geocode(context.Background(), r)
		if err != nil {
			log.Println(err)
			return utc, timezone, 0, 0, err
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
		return utc, timezone, 0, 0, err
	}

	utcOffset := (tz.RawOffset + tz.DstOffset) / 3600
	utc = strconv.Itoa(utcOffset)
	if utcOffset >= 0 {
		utc = "+" + utc
	}
	timezone = tz.TimeZoneName

	return utc, timezone, lat, lng, err
}

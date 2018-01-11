package ldapxrest

import (
	"context"
	"errors"
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

func GetAll(ldapc Connection, heads bool) (map[string]map[string]string, error) {
	var all = make(map[string]map[string]string)

	roles, err := GetRoles(ldapc)
	if err != nil {
		return all, err
	}
	var mapPeopleRole = make(map[string]string)
	for _, role := range roles {
		for _, uid := range role.Members {
			mapPeopleRole[uid] = role.Name
		}
	}

	groups, err := GetGroups(ldapc)
	if err != nil {
		return all, err
	}

	for group, groupName := range groups {

		uids, err := GetGroupMembersSlice(ldapc, group)
		if err != nil {
			return all, err
		}

		mapUIDName, mapUIDCostCenter, err := GetPeople(ldapc, uids)
		if err != nil {
			return all, err
		}

		for _, uid := range uids {

			// we don't do work for All, when need Heads only
			if heads == true {
				if _, ok := mapPeopleRole[uid]; !ok {
					// not a head, skip iteration
					continue
				}
			}

			if _, ok := all[uid]; !ok {
				var info = make(map[string]string)

				role := "Engineer"
				if _, ok := mapPeopleRole[uid]; ok {
					role = mapPeopleRole[uid]
				}
				cc := mapUIDCostCenter[uid]
				getHumanReadableRole(&role, cc)

				info["uid"] = uid
				info["name"] = mapUIDName[uid]
				info["role"] = role
				info["group"] = "tbd" // "tbd" is the hardcode to catch it later
				all[uid] = info
			}

			// In case we have one person in more than one group
			// we clone this person with another key [9:12]
			// This is useful to have it this way, as we can
			// spot folks who aren't assigned to any group
			// or assigned to multiple
			if all[uid]["group"] != "tbd" {
				var newinfo = make(map[string]string)
				for k, v := range all[uid] {
					newinfo[k] = v
					newinfo["group"] = group
					newinfo["groupName"] = groupName
				}
				all[uid+group[9:11]] = newinfo

				continue
			}

			all[uid]["group"] = group
			all[uid]["groupName"] = groupName
		}
	}

	return all, err
}

func GetHeads(ldapc Connection) (map[string]map[string]string, error) {
	return GetAll(ldapc, true)
}

func GetTimezoneInfo(ldapc Connection, uid string) (map[string]string, error) {
	var tzInfo = make(map[string]string)

	ldapLocationData, err := ldapc.GetPeopleLocationData(uid)
	if err != nil {
		return tzInfo, err
	}
	if len(ldapLocationData) != 1 {
		return tzInfo, errors.New(uid + " is the member, though was not found in ldap.")
	}
	ldapLocation := ldapLocationData[0] // safe: we have alays one item here

	remote := isRemote(ldapLocation)
	latlng := ldapLocation.GetAttributeValue("registeredAddress")
	location := ldapLocation.GetAttributeValue("rhatLocation")
	tzInfo, err = getTimeZone(latlng, location, remote)
	if err != nil {
		return nil, err
	}

	tzInfo["remote"] = strconv.FormatBool(remote)

	return tzInfo, err
}

func GetGroupMembersGeo(ldapc Connection, group string) ([]map[string]string, error) {
	uids, err := GetGroupMembersSlice(ldapc, group)
	if err != nil {
		return nil, err
	}
	var membersgeo = make([]map[string]string, len(uids))

	mapUIDName, _, err := GetPeople(ldapc, uids)
	if err != nil {
		log.Println(err)
		return membersgeo, err
	}

	for i, uid := range uids {
		tzinfo, err := GetTimezoneInfo(ldapc, uid)
		if err != nil {
			log.Println(err)
		}

		membersgeo[i] = map[string]string{
			"uid":  uid,
			"name": mapUIDName[uid],
			"lat":  tzinfo["lat"],
			"lng":  tzinfo["lng"],
		}
	}

	return membersgeo, err
}

func GetGroupMembers(ldapc Connection, group string) (map[string]*member, error) {
	var members = map[string]*member{}

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

	var mapPeopleSquad = make(map[string]string)
	squads, err := GetSquads(ldapc, group)
	if err != nil {
		return members, err
	}
	for squad, squadName := range squads {
		uids, _ := GetGroupMembersSlice(ldapc, squad)
		for _, uid := range uids {

			if mapPeopleRole[uid] == "Steward" {
				// We don't want Steward to be a member of any squad.
				// Stewards managing squads, thus members of every squad and
				// we don't want attach Steward to squad.
				// So we skip squad assignment here
				continue
			}

			mapPeopleSquad[uid] = squadName
		}
	}

	uids, err := GetGroupMembersSlice(ldapc, group)
	if err != nil {
		return members, err
	}
	ldapPeople, err := ldapc.GetPeopleFull(uids)
	if err != nil {
		return members, err
	}

	for _, man := range ldapPeople {
		uid := man.GetAttributeValue("uid")
		name := man.GetAttributeValue("cn")
		ircnick := man.GetAttributeValue("rhatNickName")
		cc := man.GetAttributeValue("rhatCostCenter")

		data := decodeNote(man.GetAttributeValue("rhatBio"))
		remote := isRemote(man)
		co := getHumanReadableLocation(man)

		role := "Engineer"
		if _, ok := mapPeopleRole[uid]; ok {
			role = mapPeopleRole[uid]
		}
		getHumanReadableRole(&role, cc)

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

	// "links" is overloaded here by one special link that called "attr",
	// we should make sure we process it properly on frontend!
	// e.g. pile:attr=contact - tells us that dfg is not real, rather contact card
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
		people, _, err := GetPeople(ldapc, role.Members)
		if err != nil {
			return head, err
		}

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

func GetPeople(ldapc Connection, uids []string) (map[string]string, map[string]string, error) {
	var people = make(map[string]string)
	var peoplecc = make(map[string]string)

	ldapPeople, err := ldapc.GetPeopleTiny(uids)
	if err != nil {
		return people, peoplecc, err
	}
	for _, ldapMan := range ldapPeople {
		uid := ldapMan.GetAttributeValue("uid")
		fullname := ldapMan.GetAttributeValue("cn")
		cc := ldapMan.GetAttributeValue("rhatCostCenter")

		people[uid] = fullname
		peoplecc[uid] = cc
	}

	return people, peoplecc, err
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

		roleMembers := cleanUids(ldapRole.GetAttributeValues("uniqueMember"))
		roleMembers = append(roleMembers, cleanUids(ldapRole.GetAttributeValues("owner"))...)

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
	groupMembers := cleanUids(ldapGroupMembers.GetAttributeValues("uniqueMember"))
	groupMembers = append(groupMembers, cleanUids(ldapGroupMembers.GetAttributeValues("owner"))...)

	squads, err := GetSquads(ldapc, group)
	if err != nil {
		return members, err
	}
	for squad := range squads {
		ldapSquadMembers, err := ldapc.GetSquadMembers(group, squad)
		if err != nil {
			return members, err
		}

		squadMembers := cleanUids(ldapSquadMembers.GetAttributeValues("uniqueMember"))
		squadMembers = append(squadMembers, cleanUids(ldapSquadMembers.GetAttributeValues("owner"))...)
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

func getHumanReadableRole(role *string, cc string) {
		switch cc {
		case "667":
			*role = *role + " [QE]"
		case "105":
			*role = "Support Delivery"
		}
}

func getHumanReadableLocation(ldapEntry *ldap.Entry) string {
	co := ldapEntry.GetAttributeValue("rhatLocation")
	re, _ := regexp.Compile(`RH - ([a-zA-Z\s]+).*`)
	place := re.FindStringSubmatch(co)
	if len(place) == 2 {
		co = ""
		if isRemote(ldapEntry) {
			co += "Remote "
		}
		co += strings.Trim(place[1], " ")              // City
		co += ", " + ldapEntry.GetAttributeValue("co") // Country
	} else {
		tmp := strings.Replace(co, "US", "USA,", 1)
		co = tmp
	}

	return co
}

func isRemote(ldapLocation *ldap.Entry) bool {
	remote := false
	if strings.ToLower(ldapLocation.GetAttributeValue("rhatOfficeLocation")) == "remote" {
		remote = true
	}
	if strings.Contains(strings.ToLower(ldapLocation.GetAttributeValue("rhatLocation")), "remote") {
		remote = true
	}

	return remote
}

func cleanUids(uids []string) []string {
	re, _ := regexp.Compile(`uid=([a-z]+)`)
	var cleanUids []string

	for _, uid := range uids {
		cleanUids = append(cleanUids, re.FindStringSubmatch(uid)[1])
	}

	return cleanUids
}

func decodeNote(note string) map[string]string {
	result := make(map[string]string)

	// accepts:
	// pile:key=value or pile:key="value value"
	// can be separated by , or space
	re, _ := regexp.Compile(`pile:(\w*=[\w:/@.-]+|\w*="[\w\s!,:/@.-]+")`)
	// TODO: take care of error here
	pile := re.FindAllStringSubmatch(note, -1)
	// TODO: code below is fragile, very fragile
	for i := range pile {
		kv := strings.Split(pile[i][1], "=")
		result[strings.Title(kv[0])] = strings.Trim(kv[1], "\"")
	}

	return result
}

func getTimeZone(latlng string, location string, remote bool) (map[string]string, error) {
	var tzInfo = make(map[string]string)

	// TODO: take this out to configuration
	gapi := os.Getenv("GAPI")
	if gapi == "" {
		return tzInfo, errors.New("GAPI environment variable is not set. Can't find timezone!")
	}
	mapsc, err := maps.NewClient(maps.WithAPIKey(gapi))
	if err != nil {
		log.Println(err)
		return tzInfo, err
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
			return tzInfo, err
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
		return tzInfo, err
	}

	utcOffset := (tz.RawOffset + tz.DstOffset) / 3600
	utc := strconv.Itoa(utcOffset)
	if utcOffset > 0 {
		utc = "+" + utc
	} else if utcOffset == 0 {
		utc = ""
	}
	timezone := tz.TimeZoneName

	tzInfo = map[string]string{
		"utcOffset": utc,
		"tzName":    timezone,
		"lat":       strconv.FormatFloat(lat, 'f', -1, 64),
		"lng":       strconv.FormatFloat(lng, 'f', -1, 64),
	}

	return tzInfo, err
}

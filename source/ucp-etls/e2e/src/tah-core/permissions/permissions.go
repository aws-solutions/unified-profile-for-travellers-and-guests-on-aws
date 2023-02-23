package permissions

import (
	"log"
	"strings"
)

const COGNITO_GROUP_PREFIX = "cloudrack_group_"
const SUPER_ADMIN_GROUP_NAME = "CLOUDRACK_SUPER_ADMIN"

//build the name of the cognito group associated to settings
func GroupName(id string) string {
	return COGNITO_GROUP_PREFIX + id
}

func IsSuperAdmin(groups []string) bool {
	log.Printf("Checking groups %v for Super admin group %s", groups, SUPER_ADMIN_GROUP_NAME)
	return HasGroup(groups, SUPER_ADMIN_GROUP_NAME)
}

func HasGroup(groups []string, groupToCheck string) bool {
	for _, group := range groups {
		if groupToCheck == group {
			return true
		}
	}
	return false
}

func GetOrgID(groups []string) string {
	for _, group := range groups {
		if strings.HasPrefix(group, COGNITO_GROUP_PREFIX) {
			return strings.Split(group, COGNITO_GROUP_PREFIX)[1]
		}
	}
	return ""
}

func HasGroupWithID(groups []string, id string) bool {
	return HasGroup(groups, GroupName(id))
}

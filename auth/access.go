package auth

import (
  "bitbucket.org/Rusty1958/shakingdog/data"
)


func IsSlemAdmin(oktaGroups []string) (bool) {
  // Checks if the admin group is in the list of supplied groups
  // from Okta
  return data.StringInSlice(oktaGroups, "shakingdog-admin-slem")
}

func IsUserAuditAdmin(oktaGroups []string) (bool) {
  // Checks if the admin group is in the list of supplied groups
  // from Okta
  return data.StringInSlice(oktaGroups, "shakingdog-admin-useraudit")
}

package auth


func IsSlemAdmin(oktaGroups []string) (bool) {
  // Checks if the admin group is in the list of supplied groups
  // from Okta

  for _, grp := range oktaGroups {
    if grp == "shakingdog-admin-slem" {
      return true
    }
  }
  return false
}

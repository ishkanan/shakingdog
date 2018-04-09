package auth


func IsAdmin(oktaGroups []string) (bool) {
  // Checks if the admin group is in the list of supplied groups
  // from Okta

  for _, grp := range oktaGroups {
    if grp == "ShakingDog-Admin" {
      return true
    }
  }
  return false
}

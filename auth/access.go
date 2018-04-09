package auth


func IsAuthorised(tenantKey string, oktaGroups []string) (bool) {
  // Checks if a tenant is in the list of supplied groups
  // from Okta

  access := false
  for _, group := range oktaGroups {
    // groups are in format "Contact-Callpicker2-{tenant key}"
    access = access || group[20:] == tenantKey
  }
  return access
}

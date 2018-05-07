package data


func StringInSlice(values []string, value string) bool {
  for i, _ := range values {
    if values[i] == value {
      return true
    }
  }
  return false
}

func IsValidDog(dog *Dog) (bool) {
  // Validates that the details of a dog are OK to save
  statuses := []string{"Affected", "Clear", "ClearByParentage", "Carrier", "CarrierByProgeny", "Unknown"}
  return len(dog.Name) > 0 &&
    StringInSlice([]string{"D", "B", "U"}, dog.Gender) &&
    StringInSlice(statuses, dog.ShakingDogStatus) &&
    StringInSlice(statuses, dog.CecsStatus)
}

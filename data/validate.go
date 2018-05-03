package data


func IsStringInSlice(value string, values []string) (bool) {
  for _, v := range values {
    if v == value {
      return true
    }
  }
  return false
}

func IsValidDog(dog *Dog) (bool) {
  // Validates that the details of a dog are acceptable

  return len(dog.Name) > 0 &&
    IsStringInSlice(dog.Gender, []string{"D", "B", "U"}) &&
    IsStringInSlice(dog.ShakingDogStatus, []string{"Affected", "Clear", "ClearByParentage", "Carrier", "CarrierByProgeny", "Unknown"}) &&
    IsStringInSlice(dog.CecsStatus, []string{"Affected", "Clear", "ClearByParentage", "Carrier", "CarrierByProgeny", "Unknown"})
}

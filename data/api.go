package data

type AuditEntry struct {
  Id int `json:"id"`
  Stamp string `json:"stamp"`
  Actor string `json:"actor"`
  Action string `json:"action"`
}

type AuditEntries struct {
  System []AuditEntry `json:"system"`
  User []AuditEntry `json:"user"`
}

type CouplesReport struct {
  Sire Dog `json:"sire"`
  Dam Dog `json:"dam"`
  Children []Dog `json:"children"`
}

type Dogs struct {
  Dogs []Dog `json:"dogs"`
}

type DogReport struct {
  Dog Dog `json:"dog"`
  FamilyAsChild *Family `json:"familyaschild"`
  FamiliesAsParent []Family `json:"familiesasparent"`
}

type ErrorMessage struct {
  Code int `json:"code"`
  Message string `json:"message"`
}

type ErrorResponse struct {
  Error *ErrorMessage `json:"error"`
}

type GenericConfirm struct {
  Result string `json:"result"`
}

type NewDog struct {
  Dog *Dog `json:"dog"`
  Sire *Dog `json:"sire"`
  Dam *Dog `json:"dam"`
}

type NewLitter struct {
  Sire Dog `json:"sire"`
  Dam Dog `json:"dam"`
  Children []Dog `json:"children"`
}

type Redirect struct {
  Location string `json:"location"`
}

type Relationships struct {
  Relationships []Relationship `json:"relationships"`
}

type TestResult struct {
  Dog TestResultDog `json:"dog"`
  Sire *Dog `json:"sire"` // pointer allows Nil value
  Dam *Dog `json:"dam"` // pointer allows Nil value
}

type TestResultDog struct {
  Id int `json:"id"`
  Name string `json:"name"`
  Gender string `json:"gender"`
  ShakingDogStatus string `json:"shakingdogstatus"`
  CecsStatus string `json:"cecsstatus"`
  OrigShakingDogStatus string `json:"origshakingdogstatus"`
  OrigCecsDogStatus string `json:"origcecsdogstatus"`
}

type UpdateDog struct {
  DogId int `json:"dogid"`
  Name string `json:"name"`
  Gender string `json:"gender"`
}

func (trd *TestResultDog) AsDataDog() (*Dog) {
  return &Dog{
    Id: trd.Id,
    Name: trd.Name,
    Gender: trd.Gender,
    ShakingDogStatus: trd.ShakingDogStatus,
    CecsStatus: trd.CecsStatus,
  }
}

package handlers

import (
  "database/sql"

  "bitbucket.org/Rusty1958/shakingdog/auth"
  "bitbucket.org/Rusty1958/shakingdog/config"
  "bitbucket.org/Rusty1958/shakingdog/data"
)

type CouplesReport struct {
  Sire data.Dog `json:"sire"`
  Dam data.Dog `json:"dam"`
  Children []data.Dog `json:"children"`
}

type Dogs struct {
  Dogs []data.Dog `json:"dogs"`
}

type DogReport struct {
  Dog data.Dog `json:"dog"`
  FamilyAsChild *data.Family `json:"familyaschild"`
  FamiliesAsParent []data.Family `json:"familiesasparent"`
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

type HandlerContext struct {
  Config *config.Config
  DBConnection *sql.DB
  Okta *auth.Okta
}

type NewLitter struct {
  Sire data.Dog `json:"sire"`
  Dam data.Dog `json:"dam"`
  Children []data.Dog `json:"children"`
}

type Redirect struct {
  Location string `json:"location"`
}

type Relationships struct {
  Relationships []data.Relationship `json:"relationships"`
}

type TestResult struct {
  Dog TestResultDog `json:"dog"`
  Sire data.Dog `json:"sire"`
  Dam data.Dog `json:"dam"`
}

type TestResultDog struct {
  Id int `json:"id"`
  Name string `json:"name"`
  Gender string `json:"gender"`
  ShakingDogStatus string `json:"shakingdogstatus"`
  CecsStatus string `json:"cecsstatus"`
  OrigShakingDogStatus bool `json:"origshakingdogstatus,omitempty"`
  OrigCecsDogStatus bool `json:"origcecsdogstatus,omitempty"`
}

func (trd *TestResultDog) AsDataDog() (*data.Dog) {
  return &data.Dog{
    Id: trd.Id,
    Name: trd.Name,
    Gender: trd.Gender,
    ShakingDogStatus: trd.ShakingDogStatus,
    CecsStatus: trd.CecsStatus,
  }
}

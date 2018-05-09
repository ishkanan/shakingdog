package data

type Dog struct {
  Id int `json:"id"`
  Name string `json:"name"`
  Gender string `json:"gender"`
  ShakingDogStatus string `json:"shakingdogstatus"`
  CecsStatus string `json:"cecsstatus"`
  ShakingDogInferOverride bool `json:"-"`
  CecsInferOverride bool `json:"-"`
}

// a family includes ALL children across ALL litters
type Family struct {
  Sire Dog `json:"sire"`
  Dam Dog `json:"dam"`
  Children []Dog `json:"children"`
}

type Relationship struct {
  SireId int `json:"sireid"`
  SireName string `json:"sirename"`
  SireShakingDogStatus string `json:"sireshakingdogstatus"`
  DamId int64 `json:"damid"`
  DamName string `json:"damname"`
  DamShakingDogStatus string `json:"damshakingdogstatus"`
  ChildId int64 `json:"childid"`
  ChildName string `json:"childname"`
  ChildShakingDogStatus string `json:"childshakingdogstatus"`
}

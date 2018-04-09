package db

import (
  "database/sql"
)

type Tenant struct {
  Id int
  Key string
  RetentionDays int
}


func GetTenantFromKey(dbConn *sql.DB, tenantKey string) (tenant *Tenant, err error) {
  tenant = &Tenant{}
  err = dbConn.QueryRow(`
    SELECT id, tenantkey, retentiondays
    FROM [dbo].[tenant]
    WHERE tenantkey = ?`,
    tenantKey,
  ).Scan(&tenant.Id, &tenant.Key, &tenant.RetentionDays)
  return
}

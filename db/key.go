package db

import (
  "database/sql"
)


func GetKey(dbConn *sql.DB, tenantKey string, year int) (keyId string, encryptedKey string, err error) {
  err = dbConn.QueryRow(`
    SELECT keyid, keydata
    FROM [dbo].[key] k
    JOIN [dbo].[tenant] t
      ON t.id = k.tenantid
    WHERE t.tenantkey = ?
      AND k.year = ?`,
    tenantKey,
    year,
  ).Scan(&keyId, &encryptedKey)
  return
}

func GenerateKeyID(dbConn *sql.DB) (keyId string, err error) {
  err = dbConn.QueryRow(
    `EXEC [dbo].[generate_keyid]`,
  ).Scan(&keyId)
  return
}

func SaveKey(dbConn *sql.DB, tenantId int, keyId string, encryptedKey string, year int) (err error) {
  _, err = dbConn.Exec(`
    INSERT INTO [dbo].[key]
    VALUES (?, ?, ?, ?)`,
    tenantId,
    keyId,
    encryptedKey,
    year,
  )
  return
}

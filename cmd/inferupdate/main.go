package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"bitbucket.org/Rusty1958/shakingdog/config"
	"bitbucket.org/Rusty1958/shakingdog/db"
)

var (
	confFile string
	dbConn *sql.DB
)


func init() {
	flag.StringVar(&confFile, "f", "", "Path to the configuration file.")
}

func main() {
	var (
		cfg *config.Config
		err error
	)

  // parse CLI arguments
	flag.Parse()
  if flag.NFlag() < 1 {
    fmt.Println("== SLEM / CECS Register (Infer Updater) ==\n")
    flag.PrintDefaults()
    return
  }

	// read in the config file
  cfg, err = config.Load(confFile)
	if err != nil {
		log.Fatalf("Error reading configuration file - %v", err)
	}

	// create DB connection
	dbConn, err = db.NewMySQLConn(
		cfg.Server.DBHost,
		cfg.Server.DBName,
		cfg.Server.DBUserName,
		cfg.Server.DBPassword,
	)
	if err != nil {
		log.Fatalf("Error establishing database connection - %v", err)
	}

	// the overall process is:
	// 1) get orphan dogs (i.e. top of relationship "tree")
	// 2) create empty recurse history
	// 2) for each orphan, CarrierByProgeny(dog, recurse history)
	// 3) for each orphan, ClearByParentage(dog, recurse history)

	// CarrierByProgeny:
	// 1) TBC by Russell as rules need clarifying

	// ClearByParentage:
	// 1) IF dog in recurse history, WARN and return
	// 2) get immediate families of orphan
	// 3) IF self and partner in ("Clear", "ClearByParentage") then for each child:
	//    - UpdateDog("ClearByParentage")
	// 4) add dog to recurse history
	// 5) for each child:
	//    - ClearByParentage(dog, recurse history)

	// UpdateDog:
	// 1) IF inferoverride is false, update status

  os.Exit(0)
}

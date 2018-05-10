package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"bitbucket.org/Rusty1958/shakingdog/config"
	"bitbucket.org/Rusty1958/shakingdog/db"
)

var (
	clearStatuses []string
	confFile string
	history []int
	labConfirmedStatuses []string
	txConn *db.Connection
)


func init() {
	flag.StringVar(&confFile, "f", "", "Path to the configuration file.")
}

func main() {
  // parse CLI arguments
	flag.Parse()
  if flag.NFlag() < 1 {
    fmt.Println("== SLEM / CECS Register (Infer Updater) ==\n")
    flag.PrintDefaults()
    return
  }

	// read in the config file
  cfg, err := config.Load(confFile)
	if err != nil {
		log.Fatalf("ERROR: Configuration file read error - %v", err)
	}

	// create DB connection
	dbConn, err := db.NewMySQLConn(
		cfg.Server.DBHost,
		cfg.Server.DBName,
		cfg.Server.DBUserName,
		cfg.Server.DBPassword,
	)
	if err != nil {
		log.Fatalf("ERROR: Database connection establish error - %v", err)
	}

  // everything is done in one transaction, with panic safety
  txConn, err = dbConn.BeginReadUncommitted(nil)
  if err != nil {
  	log.Fatalf("ERROR: Database transaction create error - %v", err)
  }
  defer txConn.Rollback()

  // status lists
	clearStatuses = []string{"Clear", "ClearByParentage"}
	labConfirmedStatuses = []string{"Affected", "Carrier", "Clear"}

	// get orphan dogs (i.e. top of relationship "tree")
	orphans, err := db.GetOrphans(txConn)
	if err != nil {
		log.Fatalf("ERROR: GetOrphans error - %v", err)
	}
	
  // for each orphan, ClearByParentage(dog)
  history = []int{}
  for i, _ := range orphans {
    log.Printf("INFO: SetClearByParentage: Processing orphan '%s'", orphans[i].Name)
    err = SetClearByParentage(&orphans[i], "└--")
    if err != nil {
      log.Fatalf("ERROR: SetClearByParentage error - %v", err)
    }
  }

	// for each orphan, CarrierByProgeny(dog)
	history = []int{}
  for i, _ := range orphans {
    log.Printf("INFO: SetCarrierByProgeny: Processing orphan '%s'", orphans[i].Name)
		err = SetCarrierByProgeny(&orphans[i], "└--")
		if err != nil {
			log.Fatalf("ERROR: SetCarrierByProgeny error - %v", err)
		}
	}

  // try commit
  err = txConn.Commit()
  if err != nil {
    log.Fatalf("ERROR: Transaction commit error - %v", err)
  }

	// all done
  os.Exit(0)
}

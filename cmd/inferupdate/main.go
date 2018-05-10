package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"bitbucket.org/Rusty1958/shakingdog/config"
	"bitbucket.org/Rusty1958/shakingdog/data"
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

	// the overall process is:

	// 1) get orphan dogs (i.e. top of relationship "tree")
	orphans, err := db.GetOrphans(txConn)
	if err != nil {
		log.Fatalf("ERROR: GetOrphans error - %v", err)
	}
	
	// 2) for each orphan, CarrierByProgeny(dog)
	history = []int{}
  for i, _ := range orphans {
		err = SetCarrierByProgeny(&orphans[i])
		if err != nil {
			log.Fatalf("ERROR: SetCarrierByProgeny error - %v", err)
		}
	}

	// 3) for each orphan, ClearByParentage(dog)
	history = []int{}
	for i, _ := range orphans {
		log.Printf("INFO: SetClearByParentage: Processing orphan '%s'", orphans[i].Name)
		err = SetClearByParentage(&orphans[i], "└--")
		if err != nil {
			log.Fatalf("ERROR: SetClearByParentage error - %v", err)
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

func SetCarrierByProgeny(dog *data.Dog) error {
	// TBC by Russell as rules need clarifying
	return nil
}

func SetClearByParentage(dog *data.Dog, logPrefix string) error {
	// already processed dog (loop detection)?
	if data.IntInSlice(history, dog.Id) {
		log.Printf("INFO: SetClearByParentage: %s Skipping as already processed", logPrefix)
		return nil
	}
	history = append(history, dog.Id)

	// get immediate families of dog
	var err error
	var families []data.Family
	if dog.Gender == "D" {
		families, err = db.GetFamiliesOfSire(txConn, dog.Id)
	} else if dog.Gender == "B" {
		families, err = db.GetFamiliesOfDam(txConn, dog.Id)
	} else {
		log.Printf("INFO: SetClearByParentage: %s Skipping as gender is '%s'", logPrefix, dog.Gender)
		return nil
	}
	if err != nil {
		return err
	}
	if len(families) == 0 {
		log.Printf("INFO: SetClearByParentage: %s Skipping as dog has no children", logPrefix)
		return nil
	}

	// update each child to ClearByParentage if:
	// 1) self AND partner are Clear/ClearByParentage, AND
	// 2) child hasn't been processed, AND
	// 3) child hasn't been lab-tested, AND
	// 4) child inferoverride flag is FALSE
	for i, _ := range families {
		family := &families[i]

		// rule #1
		if !data.StringInSlice(clearStatuses, family.Sire.ShakingDogStatus) ||
			!data.StringInSlice(clearStatuses, family.Dam.ShakingDogStatus) {
			log.Printf("INFO: SetClearByParentage: %s Skipping children update as parents are not clear.", logPrefix)
			continue
		}

		for j, _ := range family.Children {
			child := &family.Children[j]
			
			// rule #2
			if data.IntInSlice(history, child.Id) {
				log.Printf("INFO: SetClearByParentage: %s Skipping child '%s' update as already processed.", logPrefix, child.Name)
				continue
			}

			// rule #3
			if data.StringInSlice(labConfirmedStatuses, child.ShakingDogStatus) {
				log.Printf("INFO: SetClearByParentage: %s Skipping child '%s' update as has been lab-tested.", logPrefix, child.Name)
				continue
			}

			// rule #4
			if child.ShakingDogInferOverride {
				log.Printf("INFO: SetClearByParentage: %s Skipping child '%s' update as infer override flag is set.", logPrefix, child.Name)
				continue
			}
			
			// do the update
			log.Printf("INFO: SetClearByParentage: %s Updating child '%s' with old status '%s'",
				logPrefix,
				child.Name,
				child.ShakingDogStatus,
			)
			err = db.UpdateSlemStatus(txConn, child, "ClearByParentage")
			if err != nil {
				return err
			}
			child.ShakingDogStatus = "ClearByParentage"
		}
	}

	// recurse over children (as parents)
	for i, _ := range families {
		family := &families[i]
		for j, _ := range family.Children {
			child := &family.Children[j]
			log.Printf("INFO: SetClearByParentage: %s Processing child '%s'", logPrefix, child.Name)
			err = SetClearByParentage(child, "  " + logPrefix)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

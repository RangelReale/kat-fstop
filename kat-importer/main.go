package main

import (
	"flag"
	"fmt"
	"github.com/RangelReale/filesharetop/importer"
	"github.com/RangelReale/kat-fstop"
	"gopkg.in/mgo.v2"
	"log"
	"os"
)

var version = flag.Bool("version", false, "show version and exit")
var configfile = flag.String("configfile", "", "configuration file path")

func main() {
	flag.Parse()

	// output version
	if *version {
		fmt.Printf("kat-importer version %s\n", fstopimp.VERSION)
		os.Exit(0)
	}

	// create logger
	logger := log.New(os.Stderr, "", log.LstdFlags)

	// load configuration file
	config := kat.NewConfig()

	if *configfile != "" {
		logger.Printf("Loading configuration file %s", *configfile)

		err := config.Load(*configfile)
		if err != nil {
			logger.Fatal(err)
		}
	}

	// connect to mongodb
	session, err := mgo.Dial("127.0.0.1")
	if err != nil {
		log.Panic(err)
	}
	defer session.Close()

	// create and run importer
	imp := fstopimp.NewImporter(logger, session)
	imp.Database = "fstop_kat"

	// create fetcher
	fetcher := kat.NewFetcher(config)

	// import data
	err = imp.Import(fetcher)
	if err != nil {
		logger.Fatal(err)
	}

	// consolidate data
	err = imp.Consolidate("", 48)
	if err != nil {
		logger.Fatal(err)
	}

	err = imp.Consolidate("weekly", 168)
	if err != nil {
		logger.Fatal(err)
	}
}

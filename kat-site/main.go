package main

import (
	"flag"
	"fmt"
	"github.com/RangelReale/filesharetop/site"
	"gopkg.in/mgo.v2"
	"log"
	"os"
)

var version = flag.Bool("version", false, "show version and exit")

func main() {
	flag.Parse()

	// output version
	if *version {
		fmt.Printf("kat-site version %s\n", fstopsite.VERSION)
		os.Exit(0)
	}

	// connect to mongodb
	session, err := mgo.Dial("127.0.0.1")
	if err != nil {
		log.Panic(err)
	}
	defer session.Close()

	// create logger
	logger := log.New(os.Stderr, "", log.LstdFlags)

	config := fstopsite.NewConfig(13114)
	config.Title = "KAT Top"
	config.Logger = logger
	config.Session = session
	config.Database = "fstop_kat"
	config.TopId = "weekly"

	fstopsite.RunServer(config)
}

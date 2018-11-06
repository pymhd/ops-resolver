package main

import (
	"flag"
	"time"

	"database/sql"
	_ "github.com/go-sql-driver/mysql"

	log "github.com/pymhd/go-logging"
)

type DatabaseStorage map[string]*sql.DB

var (
	dbs   DatabaseStorage = make(DatabaseStorage, 0)
	cfg   *Config

	V bool
	confFile *string
)

func init() {
	confFile = flag.String("config", "./config.yaml", "Please specify YAML config file location")
	flag.BoolVar(&V, "v", false, "verbosity flag")
}

func main() {
	//parse config file, get it location from command line flags
	flag.Parse()
	
	cfg = ParseConfig(*confFile)
	//update vendor package
	//cache.SetSaveTime(cfg.CacheSyncDur)

	if V {
		log.EnableDebug()
	}
	
	log.Debugln("Application starting...")
	log.Infoln("Creating local chache obj")

	//create accountant to prevent many simultaneous sql lookup
	var cnames []string
	for _, c := range cfg.Clouds {
		cnames = append(cnames, c.Name)
	}
	MustCreateAccountant(cfg.Workers, cnames...)

	//open DB connection and assign it to global var db
	log.Infoln("Creating all MySQL connections")
	prepareDatabaseConnections()

	// Blocking!  WEB srv start
	runHTTPServer(cfg.Ip, cfg.Port, cfg.AllowNets)
}

//another file maybe  ?
func prepareDatabaseConnections() {
	for _, c := range cfg.Clouds {
		db, err := sql.Open("mysql", c.DB)
		if err != nil {
			log.Errorf("Could not connect to database in %s: %s\n", c.Name, err)
			continue
		}
		if err := db.Ping(); err != nil {
			log.Errorf("Could not ping to database in %s: %s\n", c.Name, err)
			continue
		}
		
		db.SetMaxIdleConns(0)
		db.SetConnMaxLifetime(60 * time.Second)
		log.Infof("Successfuly connected to db in %q\n", c.Name)
		dbs[c.Name] = db
	}
}

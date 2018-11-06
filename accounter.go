package main

import (
	"time"
	log "github.com/pymhd/go-logging"
)

var workers = map[string](chan bool){}

//blocks if chan is full
func bindWorker(c string) {
	for {
		select {
		case workers[c] <- true :
			log.Debugf("Recruiting 1 resolver in cloud %s\n", c)
			return
		case <- time.After(1 * time.Second):
			log.Warningf("All resolvers in cloud %s are busy. Sleeping...\n", c)
		
		}
	}
}

func unbindWorker(c string) {
        <- workers[c]
        log.Debugf("Releasing 1 resolver in cloud %s\n", c)
}


func MustCreateAccountant(max int, clouds ...string) {
	for _, c := range clouds {
		workers[c] = make(chan bool, max)
	}
}

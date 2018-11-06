package main

import (
	"github.com/gorilla/mux"
	"net"
	"net/http"
	"strings"
	log "github.com/pymhd/go-logging"
)

type Middleware func(http.HandlerFunc) http.HandlerFunc

var Logging = func(f http.HandlerFunc) http.HandlerFunc {
	//return regular http handler but with injected code
	return func(w http.ResponseWriter, r *http.Request) {
		//do some stuff
		log.Debugf("Got %s request to %s from %s\n", r.Method, r.URL.String(), r.RemoteAddr)
		//log.Debugf("Headers: \n%s", parseHeaders(r.Header))
		// Call the next middleware/handler in chain
		f(w, r)
	}
}

var MustValidateArgs = func(f http.HandlerFunc) http.HandlerFunc {
	//return regular http handler but with injected code
	return func(w http.ResponseWriter, r *http.Request) {
		IncQueryCounter(Total)

		vars := mux.Vars(r)

		id := vars["id"]
		cloud := vars["cloud"]
		resource := vars["resource"]

		defer IncQueryCounter(Error) // if success, error will be decreased by one also =)

		if dbs[cloud] == nil {
			http.Error(w, "database conn error", http.StatusInternalServerError)
			return
		}

		if _, ok := resourceRegexps[resource]; !ok {
			http.Error(w, WrongResourceTypeMessage, http.StatusBadGateway)
			return
		}

		if !resourceRegexps[resource].MatchString(id) {
			http.Error(w, WrongResourceValueMessage, http.StatusBadGateway)
			return
		}

		DecQueryCounter(Error)
		f(w, r)
	}
}

func CheckAccess(nets []*net.IPNet) Middleware {
	return func(f http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			//reqId := r.Context().Value(reqId)
			if len(nets) == 0 {
				//empty config "allow" section - allow all
				f(w, r)
				return
			}
			from := strings.Split(r.RemoteAddr, ":")[0]
			fromIP := net.ParseIP(from)
			if fromIP == nil {
				log.Errorf("Could not resolve IP addr %s\n", r.RemoteAddr)
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			for _, network := range nets {
				if network.Contains(fromIP) {
					f(w, r)
					return
				}
			}
			log.Errorf("Ip addr: %s did not match any allowed network, raising 401 Forbidden response\n", from)
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		}
	}
}

func MwManager(f http.HandlerFunc, middlewares ...Middleware) http.HandlerFunc {
	for _, m := range middlewares {
		f = m(f)
	}
	return f
}


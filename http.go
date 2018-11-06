package main

import (
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	cache "github.com/pymhd/go-simple-cache"
	log "github.com/pymhd/go-logging"
)

type netList []*net.IPNet

func runHTTPServer(ip, port string, allow netList) {

	rolton := mux.NewRouter()
	//endpints configure
	rolton.HandleFunc("/{cloud}/{resource}/{id}", MwManager(ResolveHandler, MustValidateArgs, CheckAccess(allow), Logging))
	rolton.HandleFunc("/stats", MwManager(StatsHandler, CheckAccess(allow), Logging))

	path := fmt.Sprintf("%s:%s", ip, port)
	srv := &http.Server{
		Handler:      rolton,
		Addr:         path,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}
	// run WEB srv
	log.Fatalln(srv.ListenAndServe())
}

func ResolveHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
        args := r.URL.Query()

	id := vars["id"]
        cloud := vars["cloud"]
        resource := vars["resource"]
        ckey := fmt.Sprintf("%s-%s", cloud, id)

	if resp, ok := cache.Get(ckey).(string); ok {
		log.Debugf("Found result in cache. (%s => %s)\n", id, resp)
		w.Write([]byte(resp))
		IncQueryCounter(Success)
		return
	}
	
	log.Debugln("Failed to find resource name  in cache, SQL lookup will be executed")

	// some kind of accounting cur goroutines per cloud
	bindWorker(cloud) // blocks if channel is more then max workers
	defer unbindWorker(cloud)

	var (
		name string
		code int
		err  error
	)

	switch resource {
	case ResourceInstance:
		name, code, err = resolveInstace(cloud, id)

	case ResourceVolume:
		name, code, err = resolveVolume(cloud, id)

	case ResourceSnapshot:
		name, code, err = resolveSnapshot(cloud, id)

	case ResourceIP:
		name, code, err = resolveFloatingIp(cloud, id, args)

	default:
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		IncQueryCounter(Error)
		return
	}
	if err != nil {
		text := fmt.Sprintf("%s", err)
		http.Error(w, text, code)
		IncQueryCounter(Error)
		return
	}
	if resource != ResourceIP {
		// ip addresses should not be cached
		cache.Add(ckey, name, cfg.CacheDur) 
	}
	log.Debugf("Successfuly resolved %s name by sql lookup, added  it's name in cache (%s => %s)\n", resource, id, name)
	IncQueryCounter(Success)
	w.Write([]byte(name))

}

func resolveInstace(c, id string) (string, int, error) {
	var ResolveId interface{}
	if strings.HasPrefix(id, ResourceInstance) {
		hexId := id[9:]
		instanceId, _ := strconv.ParseInt(hexId, 16, 64)
		ResolveId = instanceId
	} else {
		ResolveId = id
	}
	resp, err := lookupNova(c, ResolveId)
	if err != nil {
		log.Errorln(err)
		return "", 404, errors.New("Not Found")
	}
	return resp, 200, nil
}

func resolveVolume(c, id string) (string, int, error) {
	resp, err := lookupCinder(c, id, ResourceVolume)
	if err != nil {
		log.Errorln(err)
		return "", 404, errors.New("Not Found")
	}
	return resp, 200, nil
}

func resolveSnapshot(c, id string) (string, int, error) {
	resp, err := lookupCinder(c, id, ResourceSnapshot)
	if err != nil {
		log.Errorln(err)
		return "", 404, errors.New("Not Found")
	}
	return resp, 200, nil
}

func resolveFloatingIp(c, ip string, args url.Values) (string, int, error) {
	format := args.Get("format")
	resp, err := lookupFloatingIp(c, ip, format)
	if err != nil {
		log.Errorln(err)
		return "", 404, errors.New("Not Found")
	}
	return resp, 200, nil
}

func StatsHandler(w http.ResponseWriter, r *http.Request) {
	success, miss := cache.Stats()
	hit := float32(success) * 100 / float32(success + miss) //repcentage
	size := cache.Size()

	total, _, err := GetQueryCounter()
	
	resp := fmt.Sprintf(`{"cache_size": %d,  "queries": %d, "success": %d, "error": %d, "cache_hit_ratio": %.2f%%}`, size, total, total - err, err, hit)
	w.Write([]byte(resp))
}

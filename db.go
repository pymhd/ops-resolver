package main

import (
	"fmt"
	"errors"
	"strings"
	log "github.com/pymhd/go-logging"
)

func lookupCinder(c, id, ctx string) (string, error) {
	var q string

	switch ctx {
	case ResourceSnapshot:
		q = fmt.Sprintf(CinderQueryTmpl, SnapshotsTable)
	case ResourceVolume:
		q = fmt.Sprintf(CinderQueryTmpl, VolumesTable)
	}

	return execCinderQuery(c, q, id)

}

func lookupNova(c string, id interface{}) (string, error) {
	switch id.(type) {
	case int64:
		id, _ := id.(int64)
		return lookupInstanceByLibvirtId(c, id)
	case string:
	        id, _ := id.(string)
	        return lookupInstanceByNovaUid(c, id)
	default:
		return "", errors.New("Unsupported instance type")
	}
}

func lookupInstanceByNovaUid(c string, id interface{}) (string, error) {
	log.Debugf("Resolving instance with uuid %q  in cloud %q\n", id, c)
	return "", nil

}

func lookupInstanceByLibvirtId(c string, id int64) (string, error) {
	log.Debugf("Resolving instance %d (instance-%08x) in cloud %q\n", id, id, c)
	//get database connection
	db := dbs[c]

	//query instance display name
	var name, project string
	err := db.QueryRow(InstanceQuery, id).Scan(&name, &project)
	if err != nil {
		log.Errorf("SQL lookup failed. Error: (%s)\n", err)
		return "", errors.New("SQL lookup failed")
	}
	
	project = strings.Replace(project, ".", "_", -1)
	
	res := fmt.Sprintf("%s.%s", project, name)
	return res, nil
}

func execCinderQuery(c, q, id string) (string, error) {
	log.Debugf("Resolving cinder resource %s  in cloud %q\n", id, c)
	//get database connection
	db := dbs[c]

	//query instance display name
	var name, project string
	err := db.QueryRow(q, id).Scan(&name, &project)
	if err != nil {
		log.Errorf("SQL lookup failed. Error: (%s)\n", err)
		return "", errors.New("SQL lookup failed")
	}
	
	project = strings.Replace(project, ".", "_", -1)
	
	res := fmt.Sprintf("%s.%s", project, name)
	return res, nil
}

func lookupFloatingIp(c, ip, format string) (string, error) {
	log.Debugf("Resolving floating ip  %s  in cloud %q\n", ip, c)
	//get database connection
	db := dbs[c]
	floatingIpQuery := fmt.Sprintf(floatIpQueryTmpl, TenantIdFiledNameMap[cfg.Release[c]])

	//query instance display name
	var uuid, name, project string
	err := db.QueryRow(floatingIpQuery, ip).Scan(&project, &uuid, &name)
	if err != nil {
		log.Errorf("SQL lookup failed. Error: (%s)\n", err)
		return "", errors.New("SQL lookup failed")
	}

	project = strings.Replace(project, ".", "_", -1)
	var res string

	switch format {
	case "id":
		res = fmt.Sprintf("%s.%s", project, uuid)
	case "name":
		res = fmt.Sprintf("%s.%s", project, name)
	default:
		res = fmt.Sprintf("%s.%s.%s", project, uuid, name)
	}
	return res, nil
}

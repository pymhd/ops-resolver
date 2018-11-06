package main

import "regexp"

//type queryCorrector map[string]string

const (
	//resources to be resolved
	ResourceIP       = "ip"
	ResourceInstance = "instance"
	ResourceVolume   = "volume"
	ResourceSnapshot = "snapshot"

	//data representation types
	OutputFlag   = "format"
	OutputAsID   = "id"
	OutputAsName = "name"

	// responce patterns
	WrongResourceTypeMessage  = "wrong resource type"
	WrongResourceValueMessage = "wrong resource value"
	WrongReleaseConfValueMessage = "Panic, unsupported release"
	
	// cache keys
	QUERY     = "q"
	SUCCESS   = "s"
	ERROR     = "e"
	CACHE_HIT = "ch"

	//Openstack Releases fix
	OpenStackLiberty = "liberty"
	OpenStackMitaka  = "mitaka"
	OpenStackNewton  = "newton"
	OpenStackOcata   = "ocata"
	OpenStackPike    = "pike"
	OpenStackQueens  = "queens"

	ProjectId = "project_id"
	TenantId  = "tenant_id"

	VolumesTable   = "volumes"
	SnapshotsTable = "snapshots"

	// database Queries
	InstanceQuery = `SELECT inst.display_name, pr.name FROM nova.instances inst 
                           JOIN keystone.project pr  ON inst.project_id = pr.id 
                         WHERE inst.id = ?;`

	CinderQueryTmpl = `SELECT tmpl.display_name, pr.name FROM cinder.%s tmpl 
                             JOIN keystone.project pr ON vol.project_id = pr.id
                           WHERE vol.id = ?`

	floatIpQueryTmpl = ` SELECT pr.name, ports.device_id, inst.display_name  FROM neutron.ports as ports  
                               JOIN neutron.floatingips as f ON f.fixed_port_id = ports.id 
                               JOIN keystone.project as pr ON pr.id = ports.%s
                               JOIN nova.instances inst ON inst.uuid = ports.device_id 
                             WHERE f.floating_ip_address = ?`
)

var (
	//regexp to verify values
	validInstance  = regexp.MustCompile(`^instance-[0-9,a-f]{8}$|^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89aAbB][a-f0-9]{3}-[a-f0-9]{12}$`)
	validUUID      = regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89aAbB][a-f0-9]{3}-[a-f0-9]{12}$`)
	validIPAddress = regexp.MustCompile(`^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`)

	//gen map with regexps
	resourceRegexps = map[string]*regexp.Regexp{
		ResourceInstance: validInstance,
		ResourceVolume:   validUUID,
		ResourceSnapshot: validUUID,
		ResourceIP:       validIPAddress,
	}
	//correct some column names depends on openstack release
	TenantIdFiledNameMap = map[string]string{OpenStackLiberty: TenantId, OpenStackMitaka: TenantId,
		OpenStackNewton: ProjectId, OpenStackOcata: ProjectId,
		OpenStackPike: ProjectId, OpenStackQueens: ProjectId,
	}

	//FlotingIpPrefixes map[string][]*net.IPNet
)

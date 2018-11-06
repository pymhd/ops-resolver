package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net"
	"strings"
	"time"
)

type Config struct {
	Ip        string   `yaml:"Ip"`
	Port      string   `yaml:"Port"`
	LogFile   string   `yaml:"LogFile"`
	CacheTTL  string   `yaml:"CacheTTL"`
	CacheSync  string  `yaml:"CacheSyncTime"`
	CacheRelease string `yaml:"CacheReleaseTime"`
	Clouds    []Cloud  `yaml:"clouds"`
	Allow     []string `yaml:"AllowFrom"`
	Workers   int      `yaml:"Workers"`
	AllowNets []*net.IPNet
	CacheDur  time.Duration
	CacheSyncDur time.Duration
	CacheReleaseDur	time.Duration
	Release   map[string]string
}



type Cloud struct {
	Name    string `yaml:"name"`
	Release string `yaml:"release"`
	DB      string `yaml:"dbURL"`
}

func ParseConfig(filename string) *Config {
	cfg := &Config{}
	fb, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	if err := yaml.Unmarshal(fb, cfg); err != nil {
		panic(err)
	}
	for _, n := range cfg.Allow {
		if !strings.Contains(n, "/") {
			n += "/32"
		}
		_, network, err := net.ParseCIDR(n)
		if err != nil {
			panic(err)
		}
		cfg.AllowNets = append(cfg.AllowNets, network)
	}

	dur, err := time.ParseDuration(cfg.CacheTTL)
	if err != nil {
		cfg.CacheDur = time.Duration(24 * time.Hour)
	} else {
		cfg.CacheDur = dur
	}

	dur, err = time.ParseDuration(cfg.CacheSync)
	if err != nil {
                cfg.CacheSyncDur = time.Duration(2 * time.Hour)
        } else {
                cfg.CacheSyncDur = dur
        }
	
	r := map[string]string{}
	cfg.Release = r

	for _, c := range cfg.Clouds {
		if _, ok := TenantIdFiledNameMap[c.Release]; !ok {
			//Unsupported release
			panic(WrongReleaseConfValueMessage)
		}
		cfg.Release[c.Name] = c.Release
	}
	return cfg
}

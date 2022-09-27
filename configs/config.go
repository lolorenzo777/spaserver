// Copyright 2022 by lolorenzo77. All rights reserved.
// Use of this source code is governed by MIT licence that can be found in the LICENSE file.

/*
configs package provides a LoadConfiguration function to load a configuration for a specific environement to run the SPA web server.

The configuration is loaded from a `toml` file and becomes available in a Configuration struct.
*/
package configs

import (
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/sunraylab/verbose"
)

// if the config file is not found in the current execution directory of the application
// then LoadConfiguration look for the config file within configdir path.
// To be customized according to your server environment
const configdir = "/etc/spa/"

// A configuration loaded from a simple TOML file.
// This struct can be customized and extended with other informations as you need.
type Configuration struct {
	Environment string // the running environnement eg. dev, prod... as you wish

	SpaDir           string        `toml:"spafiledir"`         // the dir where are located the files to serve
	HttpPort         string        `toml:"http_port"`          // the spa server port
	HttpRWTimeout    time.Duration `toml:"http_rwTimeout"`     // Read and Write http timeout
	HttpIdleTimeout  time.Duration `toml:"http_idleTimeout"`   // Idle http timeout
	HttpCacheControl bool          `toml:"http_cache-control"` // Http Cache Controle, usually disable in dev environment
}

// Load TOML config file for a specified env.
//
// The config filename syntax is: "config.{environement}.toml"
//
// if environement is empty then load the `dev` one by default.
//
// Try to locate the config file within the `configs` subpath of your current directory,
// otherwise looks for the file within the configdir path defined in const here above.
//
// More about TOML format in go https://github.com/BurntSushi/toml
func LoadConfiguration(environement string) (cfg *Configuration, err error) {
	cfg = new(Configuration)
	cfg.Environment = strings.ToLower(strings.Trim(environement, " "))
	if cfg.Environment == "" {
		cfg.Environment = "dev"
	}

	//Try to locate the config file within the `configs` subpath of your current directory,
	tomldata, errf := os.ReadFile("./configs/config." + cfg.Environment + ".toml")
	if verbose.Error("LoadConfiguration", errf) != nil {

		// looks for the file within the configdir path defined in const here above.
		dir := path.Dir(configdir)
		tomldata, errf = os.ReadFile(dir + "/config." + cfg.Environment + ".toml")
		if verbose.Error("LoadConfiguration", errf) != nil {
			return nil, errf
		}
	}

	// decode the toml file within the cfg stuct
	_, errd := toml.Decode(string(tomldata), &cfg)
	if verbose.Error("LoadConfiguration", errd) != nil {
		return nil, errd
	}

	// ensure staticdir is a valid path
	cfg.SpaDir, err = filepath.Abs(cfg.SpaDir)
	if err != nil {
		log.Printf("loading server configuration fails: %v\n", err)
		return nil, err
	}

	return cfg, nil
}

package main

import (
	"os"
	"sync"

	"github.com/pborman/getopt/v2"

	"elda/config"
	"elda/def"
	"elda/events"
	"elda/log"
)

// the main part
func main() {

	// get config file path
	help := false
	getopt.HelpColumn = 0
	getopt.FlagLong(&help, "help", 'h', "Show this help")
	getopt.FlagLong(&def.Debug, "debug", 'd', "Enable debug mode")
	getopt.FlagLong(&def.ConfFile, "config", 'c', "Path to config file")
	getopt.Parse()
	if help {
		getopt.Usage()
		return
	}

	// read config
	cfg, err := config.New(def.ConfFile)
	if err != nil {
		log.Err("Config file '%s' error: %v\n", def.ConfFile, err)
		os.Exit(2)
	} else if err = cfg.Parse(); err != nil {
		log.Err("Failed to parse config file '%s': %v\n", def.ConfFile, err)
		os.Exit(3)
	}

	var wg sync.WaitGroup

	// start actions
	for name, act := range cfg.Actions() {
		if err := act.Init(); err != nil {
			log.Err("failed to init action '%s': %v\n", name, err)
			log.Err("action '%s' is disabled\n", name)
		} else {
			log.Info("starting action '%s'\n", name)
			wg.Add(1)
			a := act
			go func() { defer wg.Done(); a.Run() }()
		}
	}

	// start sources
	for name, src := range cfg.Sources() {
		if err := src.Init(); err != nil {
			log.Err("failed to init source '%s': %v\n", name, err)
			log.Err("source '%s' is disabled\n", name)
		} else {
			log.Info("starting source '%s'\n", name)
			wg.Add(1)
			s := src
			go func() { defer wg.Done(); s.Run() }()
		}
	}

	// start events
	events.Run(cfg.Sources(), cfg.Actions(), cfg.Events())

	wg.Wait()

	log.Info("exiting\n")
	return
}

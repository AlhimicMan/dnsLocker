package main

import (
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/go-yaml/yaml"
)

type Config struct {
	Nameservers    []string      `yaml:"nameservers"`
	Blocklist      []string      `yaml:"blocklist"`
	BlockHosts     []string      `yaml:"blockhosts"`
	BlockAddress4  string        `yaml:"blockAddress4"`
	BlockAddress6  string        `yaml:"blockAddress6"`
	ConfigUpdate   bool          `yaml:"configUpdate"`
	UpdateInterval time.Duration `yaml:"updateInterval"`
}

func loadConfig() (*Config, error) {
	config := &Config{}

	if _, err := os.Stat(*configFile); err != nil {
		return nil, err
	}

	data, err := ioutil.ReadFile(*configFile)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func configWatcher(configChanged chan struct{}) {
	var err error
	config, err = loadConfig()
	configChanged <- struct{}{}
	if err != nil {
		return
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	err = watcher.Add(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				log.Println("Config file updated, reload config")
				c, err := loadConfig()
				if err != nil {
					log.Println("Bad config: ", err)
				} else {
					log.Println("Config successfuly updated")
					config = c
					configChanged <- struct{}{}
				}
			}
		case err := <-watcher.Errors:
			log.Println("error:", err)
		}
	}
}

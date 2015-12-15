package main

import (
	"encoding/json"
	"io/ioutil"
)

type Configuration struct {
	BackEnd struct {
		Name            string
		BackendSettings interface{}
	}
}

func loadConfig(filePath string, configStruct *Configuration) {
	configuration, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Error("Couldn't load configuration file: ", err)
		loadConfig("tap.conf", configStruct)
	} else {
		jsErr := json.Unmarshal(configuration, &configStruct)
		if jsErr != nil {
			log.Error("Couldn't unmarshal configuration: ", jsErr)
		}
	}
}

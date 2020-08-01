package main

import (
	"encoding/json"
	"io/ioutil"
)

type Configs struct {
	Port     int `json:"port"`
	Database struct {
		Host     string `json:"host"`
		Name     string `json:"name"`
		User     string `json:"user"`
		Password string `json:"password"`
	} `json:"database"`
	TokenSecret string `json:"token_secret"`
}

var configs = loadConfigs()

func loadConfigs() Configs {
	var configs Configs
	data, _ := ioutil.ReadFile("configs.json")
	json.Unmarshal(data, &configs)
	return configs
}

package main

import (
	"encoding/json"
	"io/ioutil"

	"github.com/mjarkk/machinery-dashboard/plugin"
)

func getConfig() (plugin.Options, error) {
	var toReturn plugin.Options

	data, err := ioutil.ReadFile("./config.json")
	if err != nil {
		return toReturn, err
	}

	err = json.Unmarshal(data, &toReturn)
	return toReturn, err
}

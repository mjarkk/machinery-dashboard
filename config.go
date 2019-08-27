package main

import (
	"encoding/json"
	"io/ioutil"

	"github.com/mjarkk/machinery-dashboard/shared"
)

func getConfig() (shared.Options, error) {
	var toReturn shared.Options

	data, err := ioutil.ReadFile("./config.json")
	if err != nil {
		return toReturn, err
	}

	err = json.Unmarshal(data, &toReturn)
	return toReturn, err
}

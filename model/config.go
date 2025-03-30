package model

import (
	"encoding/json"
	"io"
	"os"
)

type config struct {
	InputDevice string `json:"input_device"`
}

func ReadConfig(path string) *config {
	conf := &config{}

	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	data, err := io.ReadAll(file)
	if err != nil {
		panic("cannot read file:" + err.Error())
	}

	json.Unmarshal(data, conf)

	return conf
}

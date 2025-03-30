package main

import (
	"github.com/lamasutra/mouse-wheel-axis/model"
)

func main() {
	conf := model.ReadConfig("config.json")

	reader, err := model.OpenDevice(conf.InputDevice)

	if err != nil {
		panic(err)
	}

	vjoy := model.CreateVjoy(reader, "Virtual joystick for side wheel")

	defer vjoy.Close()
	go vjoy.WatchRelease()
	for {
		vjoy.Write()
	}
}

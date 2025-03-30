package model

import (
	"fmt"
	"time"
)

type virtJoy struct {
	reader *deviceReader
	ui     *userInput
}

func CreateVjoy(reader *deviceReader, name string) *virtJoy {
	ui, err := CreateUserInput(name, "/dev/uinput")
	if err != nil {
		panic(err)
	}
	vjoy := &virtJoy{
		reader: reader,
		ui:     ui,
	}

	return vjoy
}

func (v *virtJoy) Close() {
	v.reader.dev.Close()
	v.ui.Close()
}

func (v *virtJoy) Write() {
	for {
		v.reader.Read()
		err := v.ui.WriteReader(v.reader)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (v *virtJoy) WatchRelease() {
	var now int64
	for {
		now = time.Now().UnixMilli()
		if v.reader.lastRelTime+50 < now && v.reader.relVal != 0 {
			v.reader.relVal = 0
			v.reader.lastRelTime = now
			err := v.ui.WriteReader(v.reader)
			if err != nil {
				fmt.Println(err)
			}
		}
		time.Sleep(time.Millisecond * 50)
	}
}

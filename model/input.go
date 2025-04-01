package model

import (
	"fmt"
	"log"
	"time"

	"github.com/holoplot/go-evdev"
)

type deviceReader struct {
	dev         *evdev.InputDevice
	minVal      int
	maxVal      int
	value       int
	relVal      int
	comVal      int
	lastRelTime int64
}

func OpenDevice(path string) (*deviceReader, error) {
	dev, err := evdev.Open(path)
	if err != nil {
		log.Fatalf("Failed to open device: %v", err)
		return nil, err
	}

	name, err := dev.Name()

	if err != nil {
		fmt.Println("Cannot get device name, we will try to use it anyway")
	} else {
		fmt.Println("Listening for events from", name)
	}

	reader := deviceReader{
		dev:    dev,
		minVal: -32767,
		maxVal: 32767,
	}

	return &reader, nil
}

func (r *deviceReader) Read() {
	event, err := r.dev.ReadOne()
	if err != nil {
		log.Fatalf("Error reading events: %v", err)
	}

	now := time.Now().UnixMilli()

	// Relative motion events
	if event.Type == evdev.EV_REL {
		var step int
		// scroll
		if event.Code == evdev.REL_HWHEEL {
			r.lastRelTime = now
			if event.Value > 0 {
				r.relVal = 1
			} else {
				r.relVal = -1
			}

			step = int(event.Value * 1024)
			r.value += step
			r.comVal += step
			if event.Value > 0 {
				if r.comVal > 0 {
					r.comVal = 0
				}
				if r.value > r.maxVal {
					r.value = r.maxVal
				}
			} else {
				if r.comVal < r.minVal {
					r.comVal = r.minVal
				}
				if r.value < r.minVal {
					r.value = r.minVal
				}
			}
			if r.comVal > 0 {
				r.comVal = 0
			}
			fmt.Println("step:", step, "value:", event.Value, "x:", r.value, "rx:", r.comVal, "xh", r.relVal)
		}
	}
}

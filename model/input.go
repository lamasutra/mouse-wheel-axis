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
		// scroll
		if event.Code == evdev.REL_HWHEEL {
			r.lastRelTime = now
			r.relVal = int(event.Value)
			if event.Value == 1 {
				r.value += 2048
				if r.value > r.maxVal {
					r.value = r.maxVal
				}
				// fmt.Printf("Scroll down")
			} else {
				r.value -= 2048
				if r.value < r.minVal {
					r.value = r.minVal
				}
				// fmt.Printf("Scroll up")
			}
			// fmt.Println("rel", r.relVal)
		}
	}
}

package model

import (
	"fmt"
	"log"
	"math"
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

func getIncrement(baseIncrement, threshold, maxValue int32, value int32) int32 {
	if value < threshold {
		return baseIncrement
	}

	// Apply an exponential decay factor
	scale := float64(maxValue - threshold) // The range where decay happens
	progress := float64(value-threshold) / scale
	decayFactor := math.Pow(0.5, progress*5) // Exponential decay

	// Ensure the increment never becomes zero
	return int32(math.Max(float64(baseIncrement)*decayFactor, 5))
}

func (r *deviceReader) Read() {
	event, err := r.dev.ReadOne()
	if err != nil {
		log.Fatalf("Error reading events: %v", err)
	}

	now := time.Now().UnixMilli()

	var baseIncrement int32 = 512
	var threshold int32 = 32000
	var maxValue int32 = 32767

	// Relative motion events
	if event.Type == evdev.EV_REL {
		var step, comStep int
		// scroll
		if event.Code == evdev.REL_HWHEEL {
			r.lastRelTime = now
			if event.Value > 0 {
				r.relVal = 1
			} else {
				r.relVal = -1
			}

			step = int(event.Value * baseIncrement)
			comStep = r.relVal * int(getIncrement(baseIncrement, threshold-baseIncrement, maxValue, int32(r.comVal)))
			r.value += step
			r.comVal += comStep
			if event.Value > 0 {
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
			if r.comVal < 16384 {
				r.comVal = 16384
			} else if r.comVal > 32767 {
				r.comVal = 32767
			}
			fmt.Println("step:", step, "value:", event.Value, "x:", r.value, "rx:", r.comVal, "xh", r.relVal)
		}
	}
}

package model

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"syscall"
	"time"
)

// Define uinput constants
const (
	MAX_NAME_LENGTH = 80

	UI_DEV_CREATE  = 0x5501
	UI_DEV_DESTROY = 0x5502
	UI_SET_EVBIT   = 0x40045564
	UI_SET_ABSBIT  = 0x40045567

	EV_SYN = 0x00
	EV_ABS = 0x03

	ABS_X  = 0x00
	ABS_HX = 0x10

	BUS_USB = 0x03

	P_VID = 0x1209
	P_PID = 0x262A
)

type userInput struct {
	dev          *uiDev
	file         *os.File
	lastValue    int
	lastRelValue int
}

type uiId struct {
	bustype uint16
	vendor  uint16
	product uint16
	version uint16
}
type uiDev struct {
	name    [80]byte
	id      uiId
	emax    uint32
	absMax  [64]int32
	absMin  [64]int32
	absFuzz [64]int32
	absFlat [64]int32
}

type uiEvent struct {
	Time  syscall.Timeval
	Type  uint16
	Code  uint16
	Value int32
}

func eventToBytes(s uiEvent) ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, 24))
	err := binary.Write(buf, binary.LittleEndian, s)
	if err != nil {
		return nil, fmt.Errorf("failed to write input event to buffer: %v", err)
	}
	return buf.Bytes(), nil
}

func ioctl(file *os.File, request uintptr, value uintptr) syscall.Errno {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, file.Fd(), request, value)
	if errno != 0 {
		return errno
	}
	return 0
}

func CreateUserInput(name string, path string) (*userInput, error) {
	// Open uinput device
	file, err := os.OpenFile(path, syscall.O_WRONLY|syscall.O_NONBLOCK, 0660)
	if err != nil {
		return nil, errors.New("Failed to open /dev/uinput:" + err.Error())
	}

	// Configure joystick device
	uidev := uiDev{
		name: toDevName(name),
		id: uiId{
			bustype: BUS_USB,
			vendor:  P_VID,
			product: P_PID,
			version: 1,
		},
	}
	uidev.absMax[ABS_X] = 32767
	uidev.absMin[ABS_X] = -32767
	uidev.absMax[ABS_HX] = 1
	uidev.absMin[ABS_HX] = -1

	// enable absolute axes
	errno := ioctl(file, UI_SET_EVBIT, uintptr(EV_ABS))
	if errno != 0 {
		return nil, errors.New("failed to allocate evbit: " + fmt.Sprint(errno))
	}

	// enable abs x
	errno = ioctl(file, UI_SET_ABSBIT, uintptr(ABS_X))
	if errno != 0 {
		return nil, errors.New("failed to reate ABS_X: " + fmt.Sprint(errno))
	}

	// enable abs hat x
	errno = ioctl(file, UI_SET_ABSBIT, uintptr(ABS_HX))
	if errno != 0 {
		return nil, errors.New("failed to reate ABS_HX: " + fmt.Sprint(errno))
	}

	createDevice(file, uidev)

	fmt.Println("Virtual Joystick `", name, "` Created.")

	ui := userInput{
		dev:  &uidev,
		file: file,
	}

	return &ui, nil
}

func toDevName(name string) [MAX_NAME_LENGTH]byte {
	var fsName [MAX_NAME_LENGTH]byte
	copy(fsName[:], []byte(name))
	return fsName
}

func createDevice(file *os.File, uidev uiDev) error {
	// errno := ioctl(file, UI_DEV_CREATE, uintptr(unsafe.Pointer(&uidev)))
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, uidev)
	if err != nil {
		_ = file.Close()
		return fmt.Errorf("failed to write user device buffer: %v", err)
	}
	_, err = file.Write(buf.Bytes())
	if err != nil {
		_ = file.Close()
		return fmt.Errorf("failed to write uidev struct to device file: %v", err)
	}

	errno := ioctl(file, UI_DEV_CREATE, uintptr(0))
	if errno != 0 {
		return errors.New("failed to create device: " + fmt.Sprint(errno))
	}

	time.Sleep(time.Millisecond * 200)

	return nil
}

func destroyDevice(file *os.File) error {
	errno := ioctl(file, UI_DEV_DESTROY, uintptr(0))
	if errno != 0 {
		return errors.New("failed to destroy device: " + fmt.Sprint(errno))
	}

	return nil
}

func (u *userInput) Close() {
	destroyDevice(u.file)
	u.file.Close()
}

func (u *userInput) WriteReader(r *deviceReader) error {
	if u.lastValue != r.value {
		err := u.Write(r.value)
		if err != nil {
			return err
		}
		u.lastValue = r.value
	}
	if u.lastRelValue != r.relVal {
		err := u.WriteRel(r.relVal)
		if err != nil {
			return err
		}
		u.lastRelValue = r.relVal
	}

	return nil
}

func (u *userInput) Write(value int) error {
	err := u.writeEvent(uiEvent{
		Type:  EV_ABS,
		Code:  ABS_X,
		Value: int32(value),
	})
	if err != nil {
		return err
	}

	return u.syncEvents()
}

func (u *userInput) WriteRel(value int) error {
	err := u.writeEvent(uiEvent{
		Type:  EV_ABS,
		Code:  ABS_HX,
		Value: int32(value),
	})
	if err != nil {
		return err
	}
	return u.syncEvents()
}

func (u *userInput) syncEvents() error {
	return u.writeEvent(uiEvent{
		Time:  syscall.Timeval{Sec: 0, Usec: 0},
		Type:  EV_SYN,
		Code:  uint16(0),
		Value: 0,
	})
}

func (u *userInput) writeEvent(ev uiEvent) error {
	buf, err := eventToBytes(ev)
	if err != nil {
		return err
	}

	_, err = u.file.Write(buf)
	if err != nil {
		return err
	}

	return nil
}

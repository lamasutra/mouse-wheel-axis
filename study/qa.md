# Q: how to read mx master 3s side wheel in go ?

Reading the side scroll wheel of the **Logitech MX Master 3S** in **Go** requires accessing the HID (Human Interface Device) events from the mouse. Since Logitech’s MX Master 3S has a horizontal scroll wheel, it typically maps to **HID events or specific scroll axes**.

### Steps to Read the Side Wheel in Go:

#### 1. **Identify the Device**
   - The MX Master 3S usually appears as a **USB HID device**.
   - On Linux, you can find it using:
     ```sh
     ls /dev/input/by-id/
     ```
   - On macOS, use:
     ```sh
     ioreg -p IOUSB
     ```

#### 2. **Use a Go HID Library**
   - The [gousb](https://github.com/google/gousb) library can be used for USB HID interaction.
   - The [go-evdev](https://github.com/gvalkov/go-evdev) library is useful for Linux-based event handling.

#### 3. **Read Input Events using go-evdev (Linux)**
   ```go
   package main

   import (
       "fmt"
       "log"
       "os"

       evdev "github.com/gvalkov/go-evdev"
   )

   func main() {
       // Open the mouse event file (may vary, check /dev/input/)
       dev, err := evdev.Open("/dev/input/eventX") // Replace eventX with correct device
       if err != nil {
           log.Fatalf("Failed to open device: %v", err)
       }

       fmt.Printf("Listening for events from %s...\n", dev.Name)

       for {
           events, err := dev.Read()
           if err != nil {
               log.Fatalf("Error reading events: %v", err)
           }

           for _, ev := range events {
               if ev.Type == evdev.EV_REL { // Relative motion events
                   if ev.Code == evdev.REL_WHEEL { // Vertical scroll
                       fmt.Printf("Vertical Scroll: %d\n", ev.Value)
                   } else if ev.Code == evdev.REL_HWHEEL { // Side scroll wheel
                       fmt.Printf("Side Scroll: %d\n", ev.Value)
                   }
               }
           }
       }
   }
   ```

#### 4. **Windows Alternative**
   - Use [github.com/karalabe/hid](https://github.com/karalabe/hid) to interface with HID devices.
   - You may need **Logi Options+ SDK** for custom inputs.

### Debugging
   - Run `evtest` on Linux to verify the device events:
     ```sh
     sudo evtest
     ```
   - Move the **side wheel** to see which event codes are triggered.

Would you like a Windows-specific implementation? 🚀

# Q: how to create virtual joystick in linux and go 

To create a virtual joystick in Linux using Go, you need to interact with the **uinput** subsystem, which allows user-space applications to create and manage input devices, including joysticks. Here's how you can do it:

---

## **1. Install Dependencies**
You'll need the `uinput` kernel module:
```sh
sudo modprobe uinput
```
Ensure it's loaded on startup:
```sh
echo "uinput" | sudo tee -a /etc/modules
```

---

## **2. Use the `go-vjoy` or `evdev` Library**
There is no native Go library for uinput, but you can use **`evdev`** to communicate with the Linux input subsystem.

### **Example: Creating a Virtual Joystick in Go**
This example creates a virtual joystick using `golang.org/x/sys/unix` for direct system calls.

#### **Install Dependencies**
```sh
go get golang.org/x/sys/unix
```

#### **Go Code: Virtual Joystick**
```go
package main

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

// Define uinput constants
const (
	UI_DEV_CREATE = 0x5501
	UI_DEV_DESTROY = 0x5502
	UI_SET_EVBIT   = 0x40045564
	UI_SET_KEYBIT  = 0x40045565
	UI_SET_ABSBIT  = 0x40045567

	EV_KEY = 0x01
	EV_ABS = 0x03

	ABS_X = 0x00
	ABS_Y = 0x01
	BTN_A = 0x130
)

func main() {
	// Open uinput device
	file, err := os.OpenFile("/dev/uinput", os.O_WRONLY|os.O_NONBLOCK, 0660)
	if err != nil {
		fmt.Println("Failed to open /dev/uinput:", err)
		return
	}
	defer file.Close()

	// Enable absolute axes and buttons
	ioctl(file.Fd(), UI_SET_EVBIT, EV_ABS)
	ioctl(file.Fd(), UI_SET_EVBIT, EV_KEY)

	ioctl(file.Fd(), UI_SET_ABSBIT, ABS_X)
	ioctl(file.Fd(), UI_SET_ABSBIT, ABS_Y)
	ioctl(file.Fd(), UI_SET_KEYBIT, BTN_A)

	// Configure joystick device
	uidev := struct {
		name      [80]byte
		id_bustype uint16
		id_vendor  uint16
		id_product uint16
		id_version uint16
		absmax     [6]int32
	}{
		id_bustype: 0x03,
		id_vendor:  0x1234,
		id_product: 0x5678,
		id_version: 1,
		absmax:     [6]int32{32767, 32767},
	}

	copy(uidev.name[:], "Go Virtual Joystick")

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, file.Fd(), UI_DEV_CREATE, uintptr(unsafe.Pointer(&uidev)))
	if errno != 0 {
		fmt.Println("Error creating device:", errno)
		return
	}

	fmt.Println("Virtual Joystick Created. Move it with inputs!")

	// Sleep to keep the device active
	select {} // Keep running
}

func ioctl(fd uintptr, request uintptr, value uintptr) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, request, value)
	if errno != 0 {
		return errno
	}
	return nil
}
```

---

## **3. Sending Joystick Events**
To send joystick events, write event structs to `/dev/uinput`:
```go
type inputEvent struct {
	Time  syscall.Timeval
	Type  uint16
	Code  uint16
	Value int32
}
```

To move the joystick:
```go
event := inputEvent{
	Type:  EV_ABS,
	Code:  ABS_X,
	Value: 10000, // Move X axis
}
syscall.Write(file.Fd(), (*byte)(unsafe.Pointer(&event)), unsafe.Sizeof(event))
```

---

## **4. Verify the Virtual Joystick**
You can check if the virtual joystick is created using:
```sh
ls /dev/input/
```
Or use `evtest`:
```sh
sudo apt install evtest
sudo evtest
```

---

## **5. Destroy the Virtual Joystick**
To remove the joystick:
```go
syscall.Syscall(syscall.SYS_IOCTL, file.Fd(), UI_DEV_DESTROY, 0)
```

---

### **Alternative Approach**
You can also use `cgo` to call C `libevdev` or `libuinput` for better abstraction.

---

Would you like a more feature-rich example, such as reading from a real joystick and emulating it? 🚀

# Q: please provide example for libevdev and libuinput

To create a virtual joystick in Linux using **libevdev** (for reading input) and **libuinput** (for creating virtual devices), you need to use **cgo** in Go. Here’s how to achieve it:

---

## **1. Install Dependencies**
You'll need the development libraries for `libevdev` and `libuinput`:

```sh
sudo apt install libevdev-dev libuinput-dev
```

---

## **2. Go Code: Using `cgo` with `libevdev` and `libuinput`**
This example does the following:
1. Creates a virtual joystick using **libuinput**.
2. Reads input events from a real joystick using **libevdev**.
3. Sends those events to the virtual joystick.

```go
package main

/*
#cgo LDFLAGS: -levdev -luinput
#include <stdio.h>
#include <fcntl.h>
#include <unistd.h>
#include <linux/uinput.h>
#include <linux/input.h>
#include <libevdev/libevdev.h>
#include <string.h>
#include <stdlib.h>

// Function to create a virtual joystick using uinput
int create_virtual_joystick() {
    int fd = open("/dev/uinput", O_WRONLY | O_NONBLOCK);
    if (fd < 0) return -1;

    ioctl(fd, UI_SET_EVBIT, EV_KEY);
    ioctl(fd, UI_SET_EVBIT, EV_ABS);
    
    ioctl(fd, UI_SET_ABSBIT, ABS_X);
    ioctl(fd, UI_SET_ABSBIT, ABS_Y);
    
    struct uinput_user_dev uidev;
    memset(&uidev, 0, sizeof(uidev));
    snprintf(uidev.name, UINPUT_MAX_NAME_SIZE, "Go Virtual Joystick");
    uidev.id.bustype = BUS_USB;
    uidev.id.vendor = 0x1234;
    uidev.id.product = 0x5678;
    uidev.id.version = 1;
    uidev.absmax[ABS_X] = 32767;
    uidev.absmax[ABS_Y] = 32767;

    write(fd, &uidev, sizeof(uidev));
    ioctl(fd, UI_DEV_CREATE);

    return fd;
}

// Function to read input events from a real joystick using libevdev
struct libevdev* open_joystick(int *fd) {
    struct libevdev *dev = NULL;
    *fd = open("/dev/input/js0", O_RDONLY);
    if (*fd < 0) return NULL;

    if (libevdev_new_from_fd(*fd, &dev) < 0) return NULL;
    
    return dev;
}

// Function to send an event to a virtual joystick
void send_event(int fd, int type, int code, int value) {
    struct input_event event;
    memset(&event, 0, sizeof(event));
    event.type = type;
    event.code = code;
    event.value = value;
    write(fd, &event, sizeof(event));
}

// Function to synchronize events
void sync_event(int fd) {
    struct input_event ev;
    memset(&ev, 0, sizeof(ev));
    ev.type = EV_SYN;
    ev.code = SYN_REPORT;
    ev.value = 0;
    write(fd, &ev, sizeof(ev));
}

*/
import "C"
import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("Creating Virtual Joystick...")
	virtualJoystick := C.create_virtual_joystick()
	if virtualJoystick < 0 {
		fmt.Println("Failed to create virtual joystick")
		return
	}
	defer C.close(virtualJoystick)

	var fd C.int
	realJoystick := C.open_joystick(&fd)
	if realJoystick == nil {
		fmt.Println("Failed to open real joystick")
		return
	}
	defer C.close(fd)

	fmt.Println("Mapping Real Joystick to Virtual Joystick...")

	for {
		var ev C.struct_input_event
		if C.libevdev_next_event(realJoystick, C.LIBEVDEV_READ_FLAG_NORMAL, &ev) == 0 {
			C.send_event(virtualJoystick, ev._type, ev.code, ev.value)
			C.sync_event(virtualJoystick)
			fmt.Printf("Sent Event - Type: %d, Code: %d, Value: %d\n", ev._type, ev.code, ev.value)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
```

---

## **3. Compile and Run**
Save the code as `virtual_joystick.go` and build it:

```sh
go build virtual_joystick.go
sudo ./virtual_joystick
```

---

## **4. Verify Virtual Joystick**
Check if the virtual joystick is created:
```sh
ls /dev/input/
```

Or test with `evtest`:
```sh
sudo evtest
```

---

### **Explanation**
1. **`create_virtual_joystick()`**: Uses `libuinput` to create a virtual joystick.
2. **`open_joystick()`**: Uses `libevdev` to read input from a real joystick.
3. **`send_event()`**: Sends input events to the virtual joystick.
4. **`sync_event()`**: Ensures events are applied.

---

## **Next Steps**
- **Map Buttons Properly:** Modify the `send_event()` function to remap button presses.
- **Support More Axes/Buttons:** Expand `ioctl()` calls to support additional joystick features.
- **GUI Control Panel:** Use Go with `GTK` or `Qt` to create a UI for controlling the virtual joystick.

Would you like additional features like force feedback support? 🎮
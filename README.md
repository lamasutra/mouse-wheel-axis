# mouse-wheel-axis
Virtual joystick for emulating axis and hat for mouse side wheel, Logitech MX Master 3S and probably others

## Usage
To get started with the mouse-wheel-axis virtual joystick, follow these steps:

1. Clone this repository and navigate to the directory in your terminal.
2. Edit the `config.json` file to set your input device (e.g., `/dev/input/event3`).
3. Run the application using `go run .`
4. The virtual joystick will be recognized by the operating system as an axis and hat device.

Note: This is a Linux-specific implementation, so please ensure you're running on a compatible platform before attempting to use it.

## Run
go run .

## Build
go build .

## Configuration
edit config.json and set your input device, e.g. /dev/input/event3

if you are unsure which one to use, install evtest utility and check you hw configuration

for `Logitech MX Masgter 3S` search for **Logitech USB Receiver Mouse**
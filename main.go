package main

import (
	"fmt"
	"os"

	"github.com/gen2brain/beeep"
	"github.com/godbus/dbus/v5"
)

func main() {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to connect to system bus:", err)
		os.Exit(1)
	}
	defer conn.Close()

	obj := conn.Object("net.reactivated.Fprint", "/net/reactivated/Fprint/Manager")
	var device dbus.ObjectPath
	err = obj.Call("net.reactivated.Fprint.Manager.GetDefaultDevice", 0).Store(&device)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to get default device:", err)
		os.Exit(1)
	}

	fmt.Println("Using device:", device)

	matchRule := fmt.Sprintf("type='signal',interface='net.reactivated.Fprint.Device',path='%s'", string(device))
	call := conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, matchRule)
	if call.Err != nil {
		fmt.Fprintln(os.Stderr, "Failed to add match rule:", call.Err)
		os.Exit(1)
	}

	fmt.Println("Listening for signals...")

	c := make(chan *dbus.Signal, 10)
	conn.Signal(c)

	for v := range c {
		if v.Name == "net.reactivated.Fprint.Device.VerifyStatus" {
			result := v.Body[0].(string)
			fmt.Printf("Verification status: %s\n", result)
			var title, message string
			if result == "verify-match" {
				title = "✔️ Fingerprint Verified"
				message = "Authentication successful."
			} else {
				title = "❌ Fingerprint Not Verified"
				message = "Authentication failed."
			}
			err := beeep.Notify(title, message, "")
			if err != nil {
				fmt.Fprintln(os.Stderr, "Failed to send notification:", err)
			}
		}
	}
}

package main

import (
	"bytes"
	_ "embed"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/emersion/go-autostart"
	"github.com/getlantern/systray"
)

//go:embed coff.png
var coff []byte

//go:embed conn.png
var conn []byte

const ShellToUse = "bash"

func Shellout(command string) (string, string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(ShellToUse, "-c", command)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func isCapsLockOn() int16 {
	out, _, _ := Shellout("xset q | grep Caps")
	out = strings.TrimSpace(out)
	out = strings.TrimPrefix(out, "00: Caps Lock:")
	out = strings.TrimSpace(out)
	outArr := strings.Split(out, " ")
	out = outArr[0]
	if out == "on" {
		return 1
	}
	return 0
}

var ci chan []byte

var app *autostart.App
var selfLocation string

func main() {
	selfLocation, _ = os.Executable()
	// prepare
	app = &autostart.App{
		Name:        "GoCAPs",
		DisplayName: "GoCAPs help you to see when CAPS is On/Off",
		Exec:        []string{"sh", "-c", selfLocation},
	}

	ci = make(chan []byte)

	go systray.Run(func() {
		systray.SetTitle("")

		if app.IsEnabled() {
			mDisableOrig := systray.AddMenuItem("Disable autostart", "Disable GoCAPS autostart")
			go func() {
				<-mDisableOrig.ClickedCh
				if err := app.Disable(); err != nil {
					log.Fatal(err)
				}
				go RestartSelf()
				systray.Quit()
			}()
		} else {
			mEnableOrig := systray.AddMenuItem("Enable autostart", "Enable Autostart GoCAPs")
			go func() {
				<-mEnableOrig.ClickedCh
				if err := app.Enable(); err != nil {
					log.Fatal(err)
				}
				go RestartSelf()
				systray.Quit()
			}()
		}
		systray.AddSeparator()

		mRestartOrig := systray.AddMenuItem("Restart", "Restart the whole app")
		go func() {
			<-mRestartOrig.ClickedCh
			go RestartSelf()
			systray.Quit()
		}()

		mQuitOrig := systray.AddMenuItem("Quit", "Quit the whole app")
		go func() {
			<-mQuitOrig.ClickedCh
			systray.Quit()
		}()

		for {
			select {
			case icon := <-ci:
				systray.SetIcon(icon)
			}
		}
	}, func() {
		os.Exit(0)
	})

	lastState := int16(-1)

	for {
		time.Sleep(20 * time.Millisecond)
		stateNow := isCapsLockOn()

		if stateNow == lastState {
			continue
		}

		if stateNow == 1 {
			ci <- conn

		} else {
			ci <- coff

		}
		lastState = stateNow
	}

}

func RestartSelf() error {
	args := os.Args
	env := os.Environ()
	return syscall.Exec(selfLocation, args, env)
}

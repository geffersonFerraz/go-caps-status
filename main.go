package main

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/geffersonFerraz/go-caps-status/autostart"
	"github.com/getlantern/systray"
)

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
var ct chan string

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
	ct = make(chan string)

	go systray.Run(func() {
		systray.SetTitle("Caps ")

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
			case text := <-ct:
				systray.SetTitle(text)
			}
		}
	}, func() {
		os.Exit(0)
	})

	lastState := int16(-1)

	for {
		time.Sleep(100 * time.Millisecond)
		stateNow := isCapsLockOn()

		if stateNow == lastState {
			continue
		}

		if stateNow == 1 {
			ci <- conn
			ct <- "Caps ON"

		} else {
			ci <- coff
			ct <- "Caps OFF"

		}
		lastState = stateNow
	}

}

func RestartSelf() error {
	args := os.Args
	env := os.Environ()
	return syscall.Exec(selfLocation, args, env)
}

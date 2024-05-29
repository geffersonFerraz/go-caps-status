package main

import (
	"bytes"
	"os/exec"
	"strings"
	"time"

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

func isCapsLockOn() bool {

	out, _, _ := Shellout("xset q | grep Caps")
	out = strings.TrimSpace(out)
	out = strings.TrimPrefix(out, "00: Caps Lock:")
	out = strings.TrimSpace(out)
	outArr := strings.Split(out, " ")
	out = outArr[0]

	return out == "on"
}

func onReady() {
	systray.SetTitle("Caps lock")
	go func() {
		for {
			select {
			case icon := <-ci:
				systray.SetIcon(icon)
			case text := <-ct:
				systray.SetTitle(text)
			}
		}
	}()
}

func onExit() {
}

var ci chan []byte
var ct chan string

func main() {
	ci = make(chan []byte)
	ct = make(chan string)
	go systray.Run(onReady, onExit)

	lastState := false
	firstRun := true

	for {
		stateNow := isCapsLockOn()
		firstRun = false
		if !firstRun && stateNow == lastState {
			continue
		}

		if stateNow {

			ci <- conn
			ct <- "Caps ON"

		} else {
			ci <- coff
			ct <- "Caps OFF"

		}
		lastState = stateNow
		time.Sleep(100 * time.Millisecond)

	}

}

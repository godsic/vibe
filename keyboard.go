package main

import "github.com/gdamore/tcell"

func processKeyboard(event *tcell.EventKey) *tcell.EventKey {
	switch {
	case event.Rune() == ' ':
		if device.IsStarted() {
			device.Stop()
		} else {
			device.Start()
		}
		return nil
	default:
		return event
	}
	return event
}

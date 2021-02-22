package main

import (
	"github.com/gdamore/tcell/v2"
)

func processKeyboard(event *tcell.EventKey) *tcell.EventKey {
	switch {
	case event.Rune() == ' ':
		if device.IsStarted() {
			device.Stop()
		} else {
			device.Start()
		}
		return nil
	case event.Key() == tcell.KeyRight:
		if event.Modifiers()&tcell.ModCtrl != 0 {
			n, t := nextTrack()
			tracklist.SetCurrentItem(n)
			app.Draw()
			fileName, err := processTrack(t)
			if err != nil {
				vibeLogger.Println(err)
			}
			err = loadFileIntoBuffer(fileName)
			if err != nil {
				vibeLogger.Println(err)
			}
		}
		return nil
	default:
		return event
	}
}

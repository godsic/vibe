package main

import (
	"strconv"

	"github.com/rivo/tview"
)

var (
	Sources = []*Source{
		NewSource("Dell XPS 13 (9343)", 1.052, 9.7, 48000, 32, "PCM"),
		NewSource("Sabaj DA3", 1.98, 3.6, 192000, 32, "PCM"),
		NewSource("Apple USB-C to 3.5mm Headphone Adapter", 1.039, 0.9, 48000, 24, "PCM"),
		NewSource("Onkyo A-9010 (TOSLINK)", 1.0, 0.09, 48000, 32, "Software"),
	}
	sourceNum int
	source    = NewAudioDevice(SOURCE)
)

func chooseSource() error {
	done := make(chan int)

	list := tview.NewList()
	for n, s := range Sources {
		list.AddItem(s.Name, "", rune(strconv.Itoa(n)[0]), func() { done <- 0 })
	}
	app.SetRoot(list, true).Draw()

	<-done
	source.Name, _ = list.GetItemText(list.GetCurrentItem())
	err := source.Set()
	if err != nil {
		vibeLogger.Fatalln(err)
	}
	return nil
}

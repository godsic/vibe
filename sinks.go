package main

import (
	"strconv"

	"github.com/rivo/tview"
)

var (
	Sinks = []*Sink{
		NewSink("AKG K514", 34.4, 116.9),
		NewSink("AKG K702", 67.0, 100.0),
		NewSink("Sennheiser HD4.30", 23.0, 116.0),
		NewSink("Sennheiser PX90", 35.3, 104.5),
		NewSink("95 db SPL at full scale", 100e3, 95),
	}

	sinkNum int
	sink    = NewAudioDevice(SINK)
)

func chooseSink() error {
	done := make(chan int)

	list := tview.NewList()
	for n, s := range Sinks {
		list.AddItem(s.Name, "", rune(strconv.Itoa(n)[0]), func() { done <- 0 })
	}
	app.SetRoot(list, true).Draw()

	<-done
	sink.Name, _ = list.GetItemText(list.GetCurrentItem())
	err := sink.Set()
	if err != nil {
		vibeLogger.Fatalln(err)
	}
	return nil
}

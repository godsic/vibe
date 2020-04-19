package main

import (
	"strconv"

	"github.com/rivo/tview"
)

var (
	Sinks = []*Sink{
		NewSink("AKG K514", 34.4, 116.9, ""),
		NewSink("AKG K514 (equalized)", 34.4, 110.0, "gain -7 equalizer 39 0.34q 7.0 equalizer 330 0.25q -3.5 equalizer 895 0.67q -3.7 equalizer 4346 0.72q 6.8 equalizer 17370 0.5q -5.8 equalizer 1486 3.73q -2.6 equalizer 3475 0.57q 2.8 equalizer 3520 1.74q -3.7 equalizer 7986 3.55q -3.0 equalizer 11715 4.55q 1.1"),
		NewSink("AKG K702", 67.0, 100.0, ""),
		NewSink("Sennheiser HD4.30", 23.0, 116.0, ""),
		NewSink("Sennheiser PX90", 35.3, 104.5, ""),
		NewSink("95 db SPL at full scale", 100e3, 95, ""),
		NewSink("Triangle Plaisir Kari", 6.0, 90.0, ""),
		NewSink("Triangle Plaisir Kari (equalized)", 6.0, 85.0, "equalizer 73.7 1.06q -7.6 equalizer 9515 1.0q -7.7"),
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

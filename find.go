package main

import (
	"github.com/rivo/tview"
)

func find() error {
	form := tview.NewForm()
	findApp := tview.NewApplication()

	form.AddInputField("", "", 0, nil, nil)
	form.AddButton("Tracks", func() {

	})
	form.AddButton("Albums", func() {

	})
	form.AddButton("Playlists", func() {

	})

	form.AddButton("Artists", func() {

	})

	form.AddButton("Cancel", func() {
		findApp.Stop()
	})

	form.SetBorder(true).SetTitle("Find on Tidal").SetTitleAlign(tview.AlignCenter)
	findApp.SetRoot(form, true)
	app.Suspend(func() { findApp.Run() })
	return nil
}

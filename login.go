package main

import (
	"github.com/rivo/tview"
)

func credentials() error {
	done := make(chan int)
	form := tview.NewForm()
	form.AddInputField("Username", "", 0, nil, nil)
	form.AddPasswordField("Password", "", 0, '*', nil)
	form.AddButton("Login", func() {
		username := form.GetFormItem(0).(*tview.InputField).GetText()
		password := form.GetFormItem(1).(*tview.InputField).GetText()

		err := session.Login(username, password)
		if err != nil {
		} else {
			done <- 0
		}
	})

	form.SetBorder(true).SetTitle("Tidal credentials").SetTitleAlign(tview.AlignCenter)
	app.SetRoot(form, true).Draw()

	<-done
	return nil
}

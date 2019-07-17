package main

import (
	"github.com/rivo/tview"
)

func credentials() error {

	form := tview.NewForm()
	form.AddInputField("Username", "", 0, nil, nil)
	form.AddPasswordField("Password", "", 0, '*', nil)
	form.AddButton("Login", func() {
		username := form.GetFormItem(0).(*tview.InputField).GetText()
		password := form.GetFormItem(1).(*tview.InputField).GetText()

		err := session.Login(username, password)
		if err != nil {
		} else {
			app.Stop()
		}
	})

	form.SetBorder(true).SetTitle("Tidal credentials").SetTitleAlign(tview.AlignCenter)

	if err := app.SetRoot(form, true).Run(); err != nil {
		panic(err)
	}

	return nil
}

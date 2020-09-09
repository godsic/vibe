package main

import (
	"github.com/pkg/browser"

	"github.com/gdamore/tcell"
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
			form.SetFieldBackgroundColor(tcell.ColorRed)
			app.Draw()
		} else {
			done <- 0
		}
		err = session.SaveSession(sessionFn)
		if err != nil {
			vibeLogger.Println(err)
		}
	})

	form.SetBorder(true).SetTitle("Tidal credentials").SetTitleAlign(tview.AlignCenter)
	app.SetRoot(form, true).Draw()

	<-done
	return nil
}

func credentials2() error {
	done := make(chan int)

	flex := tview.NewFlex()
	flex.SetDirection(tview.FlexRow)

	browser.OpenURL(session.GetOauth2URL())

	form := tview.NewForm()
	form.AddInputField("", "", 0, nil, nil)
	form.AddButton("Login", func() {
		code := form.GetFormItem(0).(*tview.InputField).GetText()

		err := session.LoginWithOauth2Code(code)
		if err != nil {
			form.SetFieldBackgroundColor(tcell.ColorRed)
			app.Draw()
		} else {
			done <- 0
		}
		err = session.SaveSession(sessionFn)
		if err != nil {
			vibeLogger.Println(err)
		}
	})

	form.SetBorder(true).SetTitle("Please Paste Tidal Authorization Code Here").SetTitleAlign(tview.AlignCenter)

	flex.AddItem(form, 0, 1, true)
	app.SetRoot(flex, true).Draw()

	<-done
	return nil
}

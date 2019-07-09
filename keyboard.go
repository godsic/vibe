package main

import (
	"os"

	"github.com/eiannone/keyboard"
)

func processKeyboard() {
	for {
		_, key, err := keyboard.GetSingleKey()
		if err != nil {
			panic(err)
		}
		if key == keyboard.KeySpace {
			device.Stop()
		}
		if key == keyboard.KeyEsc {
			os.Exit(0)
		}
	}
}

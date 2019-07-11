package main

import (
	"os"
	"time"

	"github.com/eiannone/keyboard"
)

func processKeyboard() {
	err := keyboard.Open()
	if err != nil {
		panic(err)
	}
	defer keyboard.Close()

	for {
		_, key, err := keyboard.GetSingleKey()
		if err != nil {
			panic(err)
		}
		if key == keyboard.KeySpace {
			playerCtl <- 0
		}
		if key == keyboard.KeyEsc {
			os.Exit(0)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

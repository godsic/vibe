package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gookit/color"
)

var (
	Sources = []*Source{
		NewSource("Dell XPS 13 (9343)", 1.052, 9.7, 48000, 32, "PCM"),
		NewSource("Sabaj DA3", 1.98, 3.6, 192000, 32, "PCM"),
		NewSource("Apple USB-C to 3.5mm Headphone Adapter", 1.039, 0.9, 48000, 24, "PCM"),
		NewSource("Onkyo A-9010 (TOSLINK)", 1.0, 0.09, 48000, 32, "Software"),
	}
	sourceNum int
	source    *Source
)

func chooseSource() error {
	for n, s := range Sources {
		fmt.Println(n+1, ": ", s.Name)

	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print(color.Bold.Render("Enter Source number: "))
	card, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	sourceNum, err = strconv.Atoi(strings.TrimSpace(card))
	if err != nil {
		return err
	}

	source = Sources[sourceNum-1]
	return nil
}

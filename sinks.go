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
	Sinks = []*Sink{
		NewSink("AKG K514", 34.4, 116.9),
		NewSink("AKG K702", 67.0, 100.0),
		NewSink("Sennheiser HD4.30", 23.0, 116.0),
		NewSink("Sennheiser PX90", 35.3, 104.5),
		NewSink("Triangle Plaisir Kari", 6, 97.0),
	}

	sinkNum int
	sink    *Sink
)

func chooseSink() error {
	for n, s := range Sinks {
		fmt.Println(n+1, ": ", s.Name)

	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print(color.Bold.Render("Enter Sink number: "))
	choice, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	sinkNum, err = strconv.Atoi(strings.TrimSpace(choice))
	if err != nil {
		return err
	}

	sink = Sinks[sinkNum-1]
	return nil
}

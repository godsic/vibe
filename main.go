package main

import (
	"flag"
	"log"
	"os"

	"github.com/godsic/tidalapi"
	"github.com/rivo/tview"
)

var (
	session             *tidalapi.Session
	track               = flag.Int("track", -1, "provide Tidal track ID.")
	album               = flag.Int("album", -1, "provide Tidal album ID.")
	playlist            = flag.String("playlist", "", "provide Tidal playlist ID.")
	artist              = flag.Int("artist", -1, "provide Tidal artist ID.")
	radio               = flag.Bool("radio", false, "toggle radio (works with --artist and --track")
	mqadec              = flag.Bool("mqadec", true, "toggle MQA decoding")
	mqarend             = flag.Bool("mqarend", false, "toggle MQA rendering")
	targetSpl           = flag.Float64("loudness", 75.0, "target percieved loudness in db SPL")
	shuffle             = flag.Bool("shuffle", false, "toggle shuffle mode.")
	jitter              = flag.Bool("jitter", false, "toggle jitter logging")
	search              = flag.Bool("find", false, "toggle find dialog at startup")
	processingChannel   = make(chan *tidalapi.Track, 1)
	playerChannel       = make(chan string, 1)
	playerStatusChannel = make(chan int, 1)
	TUIChannel          = make(chan int, 1)
	tracks              []*tidalapi.Track
	app                 = tview.NewApplication()
	tracklist           = tview.NewList()
	nextTrack           func() (int, *tidalapi.Track)
	vibeLogFn           = tracksPath + "/vibe.log"
	jitterLogFn         = tracksPath + "/jitter.log"
	sessionFn           = tracksPath + "session.json"
	vibeLog             *os.File
	jitterLog           *os.File
	vibeLogger          *log.Logger
	jitterLogger        *log.Logger
)

func TUI() {
	app.SetInputCapture(processKeyboard)
	if err := app.Run(); err != nil {
		panic(err)
	}
	TUIChannel <- 0
}

func main() {

	flag.Parse()

	openLogs()
	defer closeLogs()

	go jitterWatch()
	defer close(timeChannel)

	err := takeCareOfTracksFolder()
	if err != nil {
		vibeLogger.Fatal(err)
	}

	err = cleanupProcessedTracks()
	if err != nil {
		vibeLogger.Fatal(err)
	}

	os.Setenv("TCELL_TRUECOLOR", "disable")
	session = tidalapi.NewSession(tidalapi.LOSSLESS)

	go TUI()

	err = session.LoadSession(sessionFn)

	if err != nil && session.IsValid() == false {
		err = credentials()
		if err != nil {
			vibeLogger.Fatal(err)
		}
	}

	err = chooseCard()
	if err != nil {
		vibeLogger.Fatal(err)
	}
	defer closeCard()

	err = chooseSource()
	if err != nil {
		vibeLogger.Fatal(err)
	}

	err = chooseSink()
	if err != nil {
		vibeLogger.Fatal(err)
	}

	err = initSource()
	if err != nil {
		vibeLogger.Fatal(err)
	}

	if *search {
		find()
	}

	what, id := getcmdrequest()
	tracks, err = gettracks(what, id)
	if err != nil {
		vibeLogger.Fatal(err)
	}

	removeunplayabletracks()

	tracklist.SetTitle("Tracks")

	getracklist()
	app.SetRoot(tracklist, true)

	go loopovertracks()
	play()

	<-TUIChannel
}

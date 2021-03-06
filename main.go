package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"runtime/trace"

	"github.com/godsic/tidalapi"
	_ "github.com/ianlancetaylor/cgosymbolizer"
	"github.com/rivo/tview"
)

var (
	session              *tidalapi.Session
	track                = flag.Int("track", -1, "provide Tidal track ID.")
	album                = flag.Int("album", -1, "provide Tidal album ID.")
	playlist             = flag.String("playlist", "", "provide Tidal playlist ID.")
	artist               = flag.Int("artist", -1, "provide Tidal artist ID.")
	radio                = flag.Bool("radio", false, "toggle radio (works with --artist and --track")
	mqaRend              = flag.Bool("mqa-render", false, "toggle MQA rendering")
	mqaUseCache          = flag.Bool("mqa-use-cache", true, "toggle processed MQAs cache")
	mqaMode              = flag.String("mqa-mode", "off", "toggle MQA processing")
	profile              = flag.String("profile", "", "Dump runtime trace to specified file")
	targetSpl            = flag.Float64("loudness", 75.0, "target percieved loudness in db SPL")
	mixSpl               = flag.Float64("mixloudness", 75.0, "mixing loudness in db SPL")
	noiseSpl             = flag.Float64("noise", 0.0, "add white noise (negative value is with respect to the target SPL, positive - absolute SPL")
	shuffle              = flag.Bool("shuffle", false, "toggle shuffle mode.")
	jitter               = flag.Bool("jitter", false, "toggle jitter logging")
	search               = flag.Bool("find", false, "toggle find dialog at startup")
	phase                = flag.String("phase", "goldilocks", "resampler filter phase response (minimum, intermediate, archimago's goldilocks or linear)")
	distance             = flag.Float64("distance", 1.0, "distance to speakers (applies to speakers only)")
	legacyLogin          = flag.Bool("legacy-login", false, "toggle legacy, non-OAuth2 login")
	processingChannel    = make(chan *tidalapi.Track, 1)
	playerChannel        = make(chan string, 1)
	playerStatusChannel  = make(chan int, 1)
	TUIChannel           = make(chan int, 1)
	tracks               []*tidalapi.Track
	app                  = tview.NewApplication()
	tracklist            = tview.NewList()
	nextTrack            func() (int, *tidalapi.Track)
	homeDir, _           = os.UserHomeDir()
	userFolderPathSuffix = "/.vibe/"
	tracksPathSuffix     = "/tracks/"
	configPathSuffix     = "/config/"
	logPathSuffix        = "/logs/"
	userFolderPath       = filepath.Join(homeDir, userFolderPathSuffix)
	tracksPath           = filepath.Join(userFolderPath, tracksPathSuffix)
	logsPath             = filepath.Join(userFolderPath, logPathSuffix)
	configPath           = filepath.Join(userFolderPath, configPathSuffix)
	vibeLogFn            = filepath.Join(logsPath, "vibe.log")
	jitterLogFn          = filepath.Join(logsPath, "jitter.log")
	sessionFn            = filepath.Join(configPath, "session.json")
	sinkCfgFn            = filepath.Join(configPath, "sink.json")
	sourceCfgFn          = filepath.Join(configPath, "source.json")
	vibeLog              *os.File
	jitterLog            *os.File
	vibeLogger           *log.Logger
	jitterLogger         *log.Logger
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

	err := takeCareOfUserFolder()
	if err != nil {
		vibeLogger.Fatal(err)
	}

	openLogs()
	defer closeLogs()

	vibeLogger.Printf("Re-sampler filter phase response: %v (%v)\n", *phase, phaseMap[*phase])

	go jitterWatch()
	defer close(timeChannel)

	err = cleanupProcessedTracks()
	if err != nil {
		vibeLogger.Fatal(err)
	}

	os.Setenv("TCELL_TRUECOLOR", "disable")
	session = tidalapi.NewSession(tidalapi.MASTER)

	go TUI()

	err = session.LoadSession(sessionFn)

	if err != nil && session.IsValid() == false {
		if *legacyLogin {
			err = credentials()
		} else {
			err = credentials2()
		}
		if err != nil {
			vibeLogger.Fatal(err)
		}
	}

	err = chooseCard()
	if err != nil {
		vibeLogger.Fatal(err)
	}
	defer closeCard()

	err = source.Load(sourceCfgFn)
	if err != nil {
		vibeLogger.Println(err)
		err = chooseSource()
		if err != nil {
			vibeLogger.Fatal(err)
		}
		err = source.Save(sourceCfgFn)
		if err != nil {
			vibeLogger.Fatal(err)
		}
	}

	err = sink.Load(sinkCfgFn)
	if err != nil {
		vibeLogger.Println(err)
		err = chooseSink()
		if err != nil {
			vibeLogger.Fatal(err)
		}
		err = sink.Save(sinkCfgFn)
		if err != nil {
			vibeLogger.Fatal(err)
		}
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

	if *profile != "" {
		f, err := os.Create(*profile)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		trace.Start(f)
		defer trace.Stop()
	}

	play()

	<-TUIChannel
}

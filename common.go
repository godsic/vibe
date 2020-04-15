package main

import (
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/godsic/tidalapi"
)

var (
	qualityMap = map[string]string{
		tidalapi.Quality[tidalapi.LOSSLESS]: "🆩",
		tidalapi.Quality[tidalapi.MASTER]:   "🆨",
		tidalapi.Quality[tidalapi.HIGH]:     tidalapi.Quality[tidalapi.HIGH],
		tidalapi.Quality[tidalapi.LOW]:      "💩"}
	timeChannel = make(chan jitterData, 1000)
)

type jitterData struct {
	timeIn         time.Time
	timeOut        time.Time
	requestedBytes uint32
	readBytes      int
}

func year(date string) string {
	if len(date) == 0 {
		return ""
	}
	return date[0:4]
}

func takeCareOfUserFolder() (err error) {
	for _, v := range []string{configPath, tracksPath, logsPath} {
		err = os.MkdirAll(v, 0700)
		if err != nil {
			return
		}
	}
	return
}

func cleanupProcessedTracks() (err error) {
	files, err := filepath.Glob(tracksPath + "*" + processedTracksSuffix)
	if err != nil {
		return err
	}
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			return err
		}
	}
	return nil
}

func trackList() func() (int, *tidalapi.Track) {
	history := make([]int, 10)
	i := -1
	s := len(tracks)
	return func() (int, *tidalapi.Track) {
		switch {
		case *shuffle == true:
			rand.Seed(time.Now().UTC().UnixNano())
			notunique := true
			for notunique {
				i = rand.Intn(s)
				for _, v := range history {
					if v == i {
						continue
					}
				}
				notunique = false
			}
			history = append(history[1:], i)
			return i, tracks[i]
		default:
			i++
			if i < s {
				return i, tracks[i]
			}
			return -1, nil
		}
	}
}

func openLogs() {
	var err error
	vibeLog, err = os.OpenFile(vibeLogFn, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatal(err)
	}
	vibeLogger = log.New(vibeLog, "LOG:", log.LstdFlags)

	jitterLog, err = os.OpenFile(jitterLogFn, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatal(err)
	}
	jitterLogger = log.New(jitterLog, "", log.Ltime|log.Lmicroseconds)
}

func closeLogs() {
	if err := vibeLog.Close(); err != nil {
		log.Fatal(err)
	}
	if err := jitterLog.Close(); err != nil {
		log.Fatal(err)
	}
}

func miniaudioLogger(message string) {
	vibeLogger.Printf("miniaudio: %v\n", message)
}

func jitterWatch() {
	if *jitter {
		jd0 := <-timeChannel
		t0 := jd0.timeIn
		for jd := range timeChannel {
			t := jd.timeIn
			dt := t.Sub(t0)
			t0 = t
			dtCallback := jd.timeOut.Sub(jd.timeIn)
			jitterLogger.Printf("%v %v %v %v\n", dt.Nanoseconds(), dtCallback.Nanoseconds(), jd.requestedBytes, jd.requestedBytes)
		}
	}
}

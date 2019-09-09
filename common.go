package main

import (
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
)

func year(date string) string {
	if len(date) == 0 {
		return ""
	}
	return date[0:4]
}

func takeCareOfTracksFolder() (err error) {
	return os.MkdirAll(tracksPath, 0700)
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
	i := -1
	s := len(tracks)
	return func() (int, *tidalapi.Track) {
		switch {
		case *shuffle == true:
			rand.Seed(time.Now().UTC().UnixNano())
			i = rand.Intn(s)
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

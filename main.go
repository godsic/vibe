package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/godsic/tidalapi"
	"github.com/rivo/tview"
)

var (
	session             *tidalapi.Session
	track               = flag.Int("track", -1, "provide Tidal track ID.")
	album               = flag.Int("album", -1, "provide Tidal album ID.")
	playlist            = flag.String("playlist", "", "provide Tidal playlist ID.")
	mqadec              = flag.Bool("mqadec", true, "toggle MQA decoding")
	mqarend             = flag.Bool("mqarend", false, "toggle MQA rendering")
	processingChannel   = make(chan *tidalapi.Track, 1)
	playerChannel       = make(chan string, 1)
	playerStatusChannel = make(chan int, 1)
	tracks              = make([]*tidalapi.Track, 0, 10)
	app                 = tview.NewApplication()
	tracklist           = tview.NewList()
)

func loopovertracks() {
	for n, t := range tracks {
		if t.AllowStreaming {
			if t.AudioQuality == tidalapi.Quality[tidalapi.HIGH] {
				continue
			}
			tracklist.SetCurrentItem(n)
			app.Draw()
			fileName, err := processTrack(t)
			if err != nil {
				log.Println(err)
			}
			err = loadFileIntoBuffer(fileName)
			if err != nil {
				log.Println(err)
			}
			for buffer.Len() != 0 {
				time.Sleep(500 * time.Millisecond)
			}
		}
	}
}

func getracklist() {
	for _, t := range tracks {
		if t.AllowStreaming {
			if t.AudioQuality == tidalapi.Quality[tidalapi.HIGH] {
				continue
			}
			a := new(tidalapi.Album)
			err := session.Get(tidalapi.ALBUM, t.Album.Id, a)
			if err != nil {
				continue
			}
			label := fmt.Sprintf("%s\t🎤👩 %-20.20v\t💿 %-20.20v\t🎼 %-20.20v\t📅 %s\t", qualityMap[t.AudioQuality], t.Artist.Name, a.Title, t.Title, year(a.ReleaseDate))
			tracklist.AddItem(label, "", 0, func() {
				fileName, err := processTrack(tracks[tracklist.GetCurrentItem()])
				if err != nil {
					log.Println(err)
				}
				err = loadFileIntoBuffer(fileName)
				if err != nil {
					log.Println(err)
				}
			})
			app.Draw()
			time.Sleep(time.Second)
		}
	}
}

func main() {
	defer device.Uninit()

	flag.Parse()

	session = tidalapi.NewSession(tidalapi.LOSSLESS)

	err := credentials()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Logged in.")

	err = chooseCard()
	if err != nil {
		log.Fatal(err)
	}
	defer closeCard()

	err = chooseSource()
	if err != nil {
		log.Fatal(err)
	}

	err = chooseSink()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s -> %s -> %s \n", d.Name(), source.Name, sink.Name)
	fmt.Printf("Target Perceived Loudness is %.1f db\n", TARGET_SPL)
	err = initSource()
	if err != nil {
		log.Fatal(err)
	}

	switch {
	case *track > 0:
		obj := new(tidalapi.Track)
		err = session.Get(tidalapi.TRACK, *track, obj)
		if err != nil {
			log.Fatal(err)
		}
		tracks = append(tracks, obj)
		break
	case *album > 0:
		obj := new(tidalapi.Tracks)
		err = session.Get(tidalapi.ALBUMTRACKS, *album, obj)
		if err != nil {
			log.Fatal(err)
		}
		for i := range obj.Items {
			tracks = append(tracks, &(obj.Items[i]))
		}
		break
	case len(*playlist) > 0:
		obj := new(tidalapi.Tracks)
		err = session.Get(tidalapi.PLAYLISTTRACKS, *playlist, obj)
		if err != nil {
			log.Fatal(err)
		}
		for i := range obj.Items {
			tracks = append(tracks, &(obj.Items[i]))
		}
		break
	default:
		obj := new(tidalapi.TracksFavorite)
		err = session.Get(tidalapi.FAVORITETRACKS, session.User, obj)
		if err != nil {
			log.Fatal(err)
		}
		for i := range obj.Items {
			tracks = append(tracks, &(obj.Items[i].Item))
		}
		break
	}

	tracklist.SetTitle("Tracks")

	go getracklist()
	go loopovertracks()
	play()

	if err := app.SetRoot(tracklist, true).Run(); err != nil {
		panic(err)
	}

}

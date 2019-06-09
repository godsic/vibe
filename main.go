package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/godsic/tidalapi"
)

var (
	session  *tidalapi.Session
	track    = flag.Int("track", -1, "provide Tidal track ID.")
	album    = flag.Int("album", -1, "provide Tidal album ID.")
	playlist = flag.String("playlist", "", "provide Tidal playlist ID.")
)

func main() {
	flag.Parse()

	session = tidalapi.NewSession(tidalapi.LOSSLESS)
	login, password, err := credentials()
	if err != nil {
		log.Fatal(err)
	}
	err = session.Login(login, password)
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

	fmt.Printf("Target Perceived Loudness is %.1f db\n", TARGET_SPL)

	tracks := make([]*tidalapi.Track, 0, 10)

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
		for _, v := range obj.Items {
			tracks = append(tracks, &v)
		}
		break
	case len(*playlist) > 0:
		obj := new(tidalapi.Tracks)
		err = session.Get(tidalapi.PLAYLISTTRACKS, *playlist, obj)
		if err != nil {
			log.Fatal(err)
		}
		for _, v := range obj.Items {
			tracks = append(tracks, &v)
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
	playerChannel := make(chan string, 1)
	playerStatusChannel := make(chan int, 1)
	go player(playerChannel, playerStatusChannel)

	for _, t := range tracks {
		a := new(tidalapi.Album)
		err = session.Get(tidalapi.ALBUM, t.Album.Id, a)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s : %s (%s) - %s\t%s\t", t.Artist.Name, a.Title, year(a.ReleaseDate), t.Title, t.AudioQuality)

		fileName, err := processTrack(t)
		if err != nil {
			log.Fatal(err)
		}

		playerChannel <- fileName

	}

	close(playerChannel)
	<-playerStatusChannel

}
package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/godsic/tidalapi"
)

func getcmdrequest() (string, interface{}) {
	switch {
	case *track > 0 && *radio == false:
		return tidalapi.TRACK, *track
	case *track > 0 && *radio == true:
		return tidalapi.TRACKRADIO, *track
	case *album > 0:
		return tidalapi.ALBUMTRACKS, *album
	case *artist > 0 && *radio == false:
		return tidalapi.ARTISTTOPTRACKS, *artist
	case *artist > 0 && *radio == true:
		return tidalapi.ARTISTRADIO, *artist
	case len(*playlist) > 0:
		return tidalapi.PLAYLISTTRACKS, *playlist
	default:
		return tidalapi.FAVORITETRACKS, session.UserID
	}
}

func gettracks(request string, id interface{}) ([]*tidalapi.Track, error) {
	var obj interface{}
	switch request {
	case tidalapi.TRACK:
		obj = new(tidalapi.Track)
		break
	case tidalapi.TRACKRADIO:
		obj = new(tidalapi.Tracks)
		break
	case tidalapi.ALBUMTRACKS:
		obj = new(tidalapi.Tracks)
		break
	case tidalapi.ARTISTTOPTRACKS:
		obj = new(tidalapi.Tracks)
		break
	case tidalapi.ARTISTRADIO:
		obj = new(tidalapi.Tracks)
		break
	case tidalapi.PLAYLISTTRACKS:
		obj = new(tidalapi.Tracks)
		break
	case tidalapi.FAVORITETRACKS:
		obj = new(tidalapi.TracksFavorite)
		break
	default:
		return nil, errors.New("Unknown Tidal Request")
	}
	err := session.Get(request, id, obj)
	if err != nil {
		vibeLogger.Fatal(err)
	}

	ts := make([]*tidalapi.Track, 0, 10)
	switch request {
	case tidalapi.TRACK:
		ts = append(ts, obj.(*tidalapi.Track))
		break
	case tidalapi.FAVORITETRACKS:
		objs := obj.(*tidalapi.TracksFavorite)
		for i := range objs.Items {
			ts = append(ts, &(objs.Items[i].Item))
		}
		break
	default:
		objs := obj.(*tidalapi.Tracks)
		for i := range objs.Items {
			ts = append(ts, &(objs.Items[i]))
		}
		break
	}
	return ts, nil
}

func removeunplayabletracks() {
	if len(tracks) > 0 {
		ttracks := make([]*tidalapi.Track, 0, 10)
		for _, t := range tracks {
			if t.AllowStreaming {
				if t.AudioQuality == tidalapi.Quality[tidalapi.HIGH] {
					continue
				}
				ttracks = append(ttracks, t)
			}
		}
		tracks = ttracks
	}
}

func loopovertracks() {
	nextTrack = trackList()
	for n, t := nextTrack(); t != nil; n, t = nextTrack() {
		tracklist.SetCurrentItem(n)
		app.Draw()
		fileName, err := processTrack(t)
		if err != nil {
			vibeLogger.Println(err)
		}
		err = loadFileIntoBuffer(fileName)
		if err != nil {
			vibeLogger.Println(err)
		}
		for buffer.b.Len() != 0 {
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func getracklist() {
	tracklist.SetBorder(true)
	tracklist.SetTitle("Tracklist")
	tracklist.SetHighlightFullLine(true)
	for _, t := range tracks {
		info := fmt.Sprintf("  [darkslategray]in [dimgray]%v [darkslategray]by [saddlebrown]%v [darkolivegreen](%v)", t.Album.Title, t.Artist.Name, t.Copyright)
		tracklist.AddItem(t.Title, info, 0, func() {
			fileName, err := processTrack(tracks[tracklist.GetCurrentItem()])
			if err != nil {
				vibeLogger.Println(err)
			}
			err = loadFileIntoBuffer(fileName)
			if err != nil {
				vibeLogger.Println(err)
			}
		})
	}
}

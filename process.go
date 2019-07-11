package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/godsic/tidalapi"
)

type LoudnessInfo struct {
	Iin          string `json:"input_i"`
	TPin         string `json:"input_tp"`
	LRAin        string `json:"input_lra"`
	TRin         string `json:"input_thresh"`
	Iout         string `json:"output_i"`
	TPout        string `json:"output_tp"`
	LRAout       string `json:"output_lra"`
	TRout        string `json:"output_thresh"`
	NormType     string `json:"normalization_type"`
	TargetOffset string `json:"target_offset"`
}

const (
	HWVOL = iota
	SWVOL
)

const (
	tracksPathSuffix    = "/.tracks/"
	OVERLOAD_PROTECTION = -8.0
	PCM_HEADROOM        = -4.0
	TARGET_SAMPLE_RATE  = uint(48000)
	TARGET_SPL          = 75.0
)

var (
	soxArgs    = "--buffer 524288 --multi-threaded %s -t wav -b %d %s gain %+.2g rate -a -R 198 -c 4096 -p 45 -t -b 95 %d dither"
	ffmpegArgs = "-guess_layout_max 0 -y -hide_banner -i %s -af loudnorm=I=-24:LRA=14:TP=-4:print_format=json -f null /dev/null"
	volArgs    = "%s -t wav -e signed-integer -b %d %s gain %+.2g dither"
	homeDir, _ = os.UserHomeDir()
	tracksPath = homeDir + tracksPathSuffix
)

func soxResample(fname string, gain float64, src *Source) (string, error) {
	outname := fname + ".sox"
	args := fmt.Sprintf(soxArgs, fname, src.SampleBits, outname, gain, source.SampleRate)
	cmd := exec.Command("sox", strings.Split(args, " ")...)
	_, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return outname, nil
}

func ffmpegLoudnorm(fname string) (*LoudnessInfo, error) {
	loudnessInfo := new(LoudnessInfo)
	var outBytes []byte

	fnameLoud := fname + ".json"
	outBytes, err := ioutil.ReadFile(fnameLoud)
	if err != nil {
		args := fmt.Sprintf(ffmpegArgs, fname)
		cmd := exec.Command("ffmpeg", strings.Split(args, " ")...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return nil, err
		}

		outStr := string(out)
		outStrs := strings.Split(outStr, "\n")
		outStr = strings.Join(outStrs[len(outStrs)-13:], "\n")
		outBytes = []byte(outStr)
	}

	err = json.Unmarshal(outBytes, loudnessInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	outBytes, err = json.Marshal(loudnessInfo)
	if err != nil {
		log.Println(err)
	}
	err = ioutil.WriteFile(fnameLoud, outBytes, 0640)
	if err != nil {
		log.Println(err)
	}

	return loudnessInfo, nil
}

func applyGain(fname string, gain float64, src *Source) (string, error) {
	outname := fname + ".final.wav"
	args := fmt.Sprintf(volArgs, fname, src.SampleBits, outname, gain)
	cmd := exec.Command("sox", strings.Split(args, " ")...)
	_, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return outname, nil
}

func downloadTrack(t *tidalapi.Track) (string, error) {
	fname := tracksPath + strconv.Itoa(t.ID)

	_, err := os.Stat(fname)
	if !os.IsNotExist(err) {
		return fname, nil
	}

	path := new(tidalapi.TrackPath)
	err = session.Get(tidalapi.TRACKURL, t.ID, path)
	if err != nil {
		return "", err
	}

	resp, err := http.Get(path.Url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	f, err := os.Create(fname)
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return "", err
	}
	return fname, nil
}

func CanMQADecode() bool {
	_, err := exec.LookPath("mqadec")
	if err != nil {
		return false
	}
	return true
}

func CanMQARender() bool {
	_, err := exec.LookPath("mqarender")
	if err != nil {
		return false
	}
	return true
}

func MQADecode(fname string) (string, error) {
	if *mqadec == false {
		return fname, nil
	}

	if !CanMQADecode() {
		return fname, errors.New("MQA Decoder is not avaliable")
	}

	// fmt.Printf("MQA Decoding\t")

	outname := fname + ".mqadecoded"

	cmd := exec.Command("mqadec", fname, outname)
	_, err := cmd.CombinedOutput()
	if err != nil {
		return fname, err
	}
	return outname, nil
}

func MQARender(fname string) (string, error) {
	if *mqarend == false {
		return fname, nil
	}

	if !CanMQARender() {
		return fname, errors.New("MQA Renderer is not avaliable")
	}

	// fmt.Printf("MQA Rendering\t")

	outname := fname + ".mqarendered"

	cmd := exec.Command("mqarender", fname, outname)
	_, err := cmd.CombinedOutput()
	if err != nil {
		return fname, err
	}
	return outname, nil
}

func getGain(src *Source, sink *Sink, loudness *LoudnessInfo) float64 {
	rtot := src.Rl + sink.R
	vl := src.Vout * (sink.R / rtot)
	splMax := sink.Sensitivity + 20.*math.Log10(vl)
	targetSplRel := TARGET_SPL - splMax
	Iin, _ := strconv.ParseFloat(loudness.Iin, 64)
	gain := math.Round(targetSplRel - Iin)
	return gain
}

func processTrack(t *tidalapi.Track) (string, error) {

	fname, err := downloadTrack(t)
	if err != nil {
		return "", err
	}

	if t.AudioQuality == tidalapi.Quality[tidalapi.MASTER] {
		fname, _ = MQADecode(fname)
		defer os.Remove(fname)
		fname, _ = MQARender(fname)
		defer os.Remove(fname)
	}

	loud, err := ffmpegLoudnorm(fname)
	if err != nil {
		return "", err
	}

	gain := getGain(source, sink, loud)

	outname, err := soxResample(fname, gain, source)
	if err != nil {
		return "", err
	}
	// defer os.Remove(fname)

	fmt.Printf("🎚 %.1f db\tØ %s db\t⟝ %s db ⟞\t⛰ %s db\n", gain, loud.Iin, loud.LRAin, loud.TPin)

	// outname, err := applyGain(fname, gain, source)
	// if err != nil {
	// 	return "", err
	// }

	return outname, nil
}

func processTracks() {
	for _, t := range tracks {
		if t.AllowStreaming {
			if t.AudioQuality == tidalapi.Quality[tidalapi.HIGH] {
				continue
			}
			a := new(tidalapi.Album)
			err := session.Get(tidalapi.ALBUM, t.Album.Id, a)
			if err != nil {
				log.Println(err)
				continue
			}

			fmt.Printf("%s\t🎤👩 %-20.20v\t💿 %-20.20v\t🎼 %-20.20v\t📅 %s\t", qualityMap[t.AudioQuality], t.Artist.Name, a.Title, t.Title, year(a.ReleaseDate))

			fileName, err := processTrack(t)
			if err != nil {
				log.Println(err)
				continue
			}

			playerChannel <- fileName
		}
	}
	close(playerChannel)
}

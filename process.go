package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/godsic/tidalapi"
)

type LoudnessInfo struct {
	I   float64 `json:"i"`
	TP  float64 `json:"tp"`
	LRA float64 `json:"lra"`
	TR  float64 `json:"thresh"`
	DR  float64 `json:"dr"`
}

const (
	HWVOL = iota
	SWVOL
)

const (
	tracksPathSuffix      = "/.tracks/"
	noiseFilename         = "noise.wav"
	noiseSuffix           = ".noise"
	noisedTracksSuffix    = ".noised"
	processedTracksSuffix = ".sox"
)

var (
	phaseMap = map[string]int{
		"minimum":      0,
		"intermediate": 25,
		"goldilocks":   45,
		"linear":       50,
	}
)

var (
	soxArgs      = "--buffer 524288 --multi-threaded %s -t wav -e signed-integer -b %d %s gain %+.2g rate -a -d 33 -p %d -t -b 92 %d %s"
	ffmpegArgs   = "-guess_layout_max 0 -y -hide_banner -i %s -filter_complex ebur128=peak=true -f null -"
	noiseArgs    = "-G %s -b 32 -e float -t wav %s synth whitenoise gain %+.2g"
	mixNoiseArgs = "-G -m %s %s -b 32 -e float -t wav %s"
	drArgs       = "-hide_banner -i %s -af drmeter -f null /dev/null"
	homeDir, _   = os.UserHomeDir()
	tracksPath   = homeDir + tracksPathSuffix
)

var ffmpegOutputFmt = "  Integrated loudness:\n" +
	"    I:         %f LUFS\n" +
	"    Threshold: %f LUFS\n\n" +
	"  Loudness range:\n" +
	"    LRA:       %f LU\n" +
	"    Threshold: %f LUFS\n" +
	"    LRA low:   %f LUFS\n" +
	"    LRA high:  %f LUFS\n\n" +
	"  True  peak:\n" +
	"    Peak:      %f dBFS"

func soxResample(fname string, gain float64, src *Source, sink *Sink) (string, error) {
	outname := fname + processedTracksSuffix

	_, err := os.Stat(outname)
	if !os.IsNotExist(err) {
		return outname, nil
	}

	SoxArgs := ""

	if len(sink.SoxArgs) > 0 {
		SoxArgs = sink.SoxArgs + " dither"
	} else {
		SoxArgs = "dither"
	}

	args := fmt.Sprintf(soxArgs, fname, src.SampleBits, outname, gain, phaseMap[*phase], src.SampleRate, SoxArgs)
	vibeLogger.Println(args)
	cmd := exec.Command("sox", strings.Split(args, " ")...)
	_, err = cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return outname, nil
}

func generateWhiteNoise(fname string, gain float64) (string, error) {
	outname := fname + noiseSuffix

	_, err := os.Stat(outname)
	if !os.IsNotExist(err) {
		return outname, nil
	}

	args := fmt.Sprintf(noiseArgs, fname, outname, gain)
	vibeLogger.Println(args)
	cmd := exec.Command("sox", strings.Split(args, " ")...)
	msg, err := cmd.CombinedOutput()
	if err != nil {
		vibeLogger.Println(string(msg))
		return "", err
	}
	return outname, err
}

func addWhiteNoise(fname, noiseFname string) (string, error) {
	outname := fname + noisedTracksSuffix

	_, err := os.Stat(outname)
	if !os.IsNotExist(err) {
		return outname, nil
	}

	args := fmt.Sprintf(mixNoiseArgs, fname, noiseFname, outname)
	vibeLogger.Println(args)
	cmd := exec.Command("sox", strings.Split(args, " ")...)
	msg, err := cmd.CombinedOutput()
	if err != nil {
		vibeLogger.Println(string(msg))
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

		dummy := float64(0.0)
		_, err = fmt.Sscanf(outStr, ffmpegOutputFmt,
			&(loudnessInfo.I),
			&(loudnessInfo.TR),
			&(loudnessInfo.LRA),
			&dummy,
			&dummy,
			&dummy,
			&(loudnessInfo.TP))
		if err != nil {
			vibeLogger.Println(err)
			return nil, err
		}

		args = fmt.Sprintf(drArgs, fname)
		cmd = exec.Command("ffmpeg", strings.Split(args, " ")...)
		out, err = cmd.CombinedOutput()
		if err != nil {
			return nil, err
		}

		outStr = string(out)
		outStrs = strings.Split(outStr, "\n")
		outStr = strings.Join(outStrs[len(outStrs)-2:], "\n")

		outStrs = strings.Fields(outStr)
		outStr = outStrs[len(outStrs)-1]

		loudnessInfo.DR, err = strconv.ParseFloat(outStr, 64)
		if err != nil {
			vibeLogger.Println(outStr, err)
			return nil, err
		}
	} else {
		err = json.Unmarshal(outBytes, loudnessInfo)
		if err != nil {
			vibeLogger.Println(err)
			return nil, err
		}
	}

	outBytes, err = json.Marshal(loudnessInfo)
	if err != nil {
		vibeLogger.Println(err)
	}
	err = ioutil.WriteFile(fnameLoud, outBytes, 0640)
	if err != nil {
		vibeLogger.Println(err)
	}

	return loudnessInfo, nil
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

func getGain(src *Source, sink *Sink, loudness *LoudnessInfo) (float64, float64) {
	rtot := src.Rl + sink.R
	vl := src.Vout * (sink.R / rtot)
	splMax := sink.Sensitivity + 20.*math.Log10(vl)
	targetSplRel := *targetSpl - splMax
	distanceLoss := 20. * math.Log10(*distance)
	gain := math.Round(targetSplRel - loudness.I + distanceLoss)
	return gain, splMax
}

func processTrack(t *tidalapi.Track) (string, error) {

	fname, err := downloadTrack(t)
	if err != nil {
		return "", err
	}

	if t.AudioQuality == tidalapi.Quality[tidalapi.MASTER] {
		decFname, _ := MQADecode(fname)
		defer os.Remove(decFname)
		rendFname, _ := MQARender(decFname)
		defer os.Remove(rendFname)
		fname = rendFname

	}

	loud, err := ffmpegLoudnorm(fname)
	if err != nil {
		return "", err
	}

	gain, splMax := getGain(source.dev.(*Source), sink.dev.(*Sink), loud)

	if *noiseSpl != 0.0 {

		noiseGain := 0.0

		switch {
		case *noiseSpl < 0.0:
			noiseGain = loud.I + *noiseSpl
			break
		case *noiseSpl > 0.0:
			noiseGain = *noiseSpl - splMax - gain
			break
		}

		noiseFname, err := generateWhiteNoise(fname, noiseGain)
		if err != nil {
			return "", err
		}
		defer os.Remove(noiseFname)

		noisedFname, err := addWhiteNoise(fname, noiseFname)
		if err != nil {
			return "", err
		}
		defer os.Remove(noisedFname)
		fname = noisedFname

		gain += 6.0
	}
	outname, err := soxResample(fname, gain, source.dev.(*Source), sink.dev.(*Sink))
	if err != nil {
		return "", err
	}

	return outname, nil
}

func processTracks() {
	for {
		fileName, err := processTrack(<-processingChannel)
		if err != nil {
			vibeLogger.Println(err)
			continue
		}

		playerChannel <- fileName
	}
}

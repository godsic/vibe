package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/gen2brain/malgo"
	"github.com/gookit/color"
	wav "github.com/youpy/go-wav"
)

var (
	device       = new(malgo.Device)
	deviceConfig = malgo.DefaultDeviceConfig()
)

func bitsPerSampleToDeviceFormat(bitsPerSample int) malgo.FormatType {
	switch bitsPerSample {
	case 24:
		return malgo.FormatS24
	case 32:
		return malgo.FormatS32
	case 16:
		return malgo.FormatS16
	default:
		return malgo.FormatUnknown
	}
}

var (
	cardNum int
	ctx     *malgo.AllocatedContext
	d       malgo.DeviceInfo
)

func closeCard() {
	err := ctx.Uninit()
	if err != nil {
		log.Println(err)
	}
	ctx.Free()
}
func chooseCard() error {
	backends := []malgo.Backend{malgo.BackendAlsa}

	var err error

	ctx, err = malgo.InitContext(backends, malgo.ContextConfig{}, func(message string) {
		fmt.Printf("LOG <%v>\n", message)
	})
	if err != nil {
		return err
	}

	devices, _ := ctx.Devices(malgo.Playback)

	for n, d := range devices {
		fmt.Println(n+1, ": ", d.Name())

	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print(color.Bold.Render("Enter Card number: "))
	card, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	cardNum, err = strconv.Atoi(strings.TrimSpace(card))
	if err != nil {
		return err
	}

	d = devices[cardNum-1]

	return nil
}

func playBuffer(buffer *bytes.Buffer, NChannels, BitsPerSample, SampleRate int) error {

	var err error

	channels := uint32(NChannels)
	deviceConfig.Alsa.NoMMap = 1
	deviceConfig.ShareMode = malgo.Exclusive
	deviceConfig.PerformanceProfile = malgo.Conservative

	deviceConfig.BufferSizeInMilliseconds = 1000
	deviceConfig.Channels = channels
	deviceConfig.SampleRate = uint32(SampleRate)
	deviceConfig.Format = bitsPerSampleToDeviceFormat(BitsPerSample)

	sampleSize := uint32(malgo.SampleSizeInBytes(deviceConfig.Format))
	// This is the function that's used for sending more data to the device for playback.
	onSendSamples := func(frameCount uint32, samples []byte) uint32 {
		n, _ := buffer.Read(samples)
		frameGot := uint32(n) / channels / sampleSize
		return frameGot
	}

	deviceCallbacks := malgo.DeviceCallbacks{
		Send: onSendSamples,
	}

	device, err = malgo.InitDevice(ctx.Context, malgo.Playback, &d.ID, deviceConfig, deviceCallbacks)
	if err != nil {
		log.Fatal(err)
	}
	defer device.Uninit()

	err = device.Start()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Scanln()

	return nil
}

func playFile(fname string) (err error) {
	f, err := os.Open(fname)
	if err != nil {
		return err
	}
	defer f.Close()

	r := wav.NewReader(f)
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	format, err := r.Format()
	if err != nil {
		return err
	}
	buffer := bytes.NewBuffer(data)
	// fmt.Println(int(format.NumChannels), int(format.BitsPerSample), int(format.SampleRate))
	playBuffer(buffer, int(format.NumChannels), int(format.BitsPerSample), int(format.SampleRate))
	return nil
}

func player(in chan string, status chan int) {
	for t := range in {
		err := playFile(t)
		if err != nil {
			log.Fatal(err)
		}
		defer os.Remove(t)
	}
	close(status)
}

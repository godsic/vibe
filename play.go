package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/gen2brain/malgo"
	"github.com/rivo/tview"
	wav "github.com/youpy/go-wav"
)

var (
	device       = new(malgo.Device)
	deviceConfig = malgo.DefaultDeviceConfig()
	playerCtl    = make(chan int)
	buffer       bytes.Buffer
	bufferMutex  sync.Mutex
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
	backends := []malgo.Backend{malgo.BackendAlsa, malgo.BackendWasapi}

	var err error

	ctx, err = malgo.InitContext(backends, malgo.ContextConfig{}, func(message string) {
		fmt.Printf("LOG <%v>\n", message)
	})
	if err != nil {
		return err
	}

	devices, _ := ctx.Devices(malgo.Playback)

	done := make(chan int)

	list := tview.NewList()
	for n, dd := range devices {
		// fmt.Println(n+1, ": ", s.Name)
		list.AddItem(dd.Name(), "", rune(strconv.Itoa(n)[0]), func() { done <- 0 })
	}
	app.SetRoot(list, true).Draw()

	<-done
	d = devices[list.GetCurrentItem()]
	return nil
}

func initSource() (err error) {

	channels := uint32(2)
	deviceConfig.Alsa.NoMMap = 1
	deviceConfig.ShareMode = malgo.Exclusive
	deviceConfig.PerformanceProfile = malgo.LowLatency

	deviceConfig.BufferSizeInMilliseconds = 300 * 1000
	deviceConfig.Channels = channels
	deviceConfig.SampleRate = uint32(source.SampleRate)
	deviceConfig.Format = source.SampleFormat

	sampleSize := uint32(malgo.SampleSizeInBytes(deviceConfig.Format))
	// This is the function that's used for sending more data to the device for playback.
	onSendSamples := func(frameCount uint32, samples []byte) uint32 {
		n, err := buffer.Read(samples)
		if err == io.EOF {
			return 0
		}
		frameGot := uint32(n) / channels / sampleSize
		return frameGot
	}

	deviceCallbacks := malgo.DeviceCallbacks{
		Send: onSendSamples,
	}

	device, err = malgo.InitDevice(ctx.Context, malgo.Playback, &d.ID, deviceConfig, deviceCallbacks)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func play() error {

	err := device.Start()
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func pause() error {
	if device.IsStarted() {
		err := device.Stop()
		if err != nil {
			log.Fatal(err)
		}
	}
	return nil
}

func loadFileIntoBuffer(fname string) error {
	f, err := os.Open(fname)
	if err != nil {
		return err
	}
	defer f.Close()

	r := wav.NewReader(f)

	newbuffer := new(bytes.Buffer)
	newbuffer.ReadFrom(r)
	// buffer.Reset()
	bufferMutex.Lock()
	buffer = *newbuffer
	bufferMutex.Unlock()
	return nil
}

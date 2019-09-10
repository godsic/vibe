package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/godsic/malgo"
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
	backends := []malgo.Backend{malgo.BackendAlsa, malgo.BackendWasapi, malgo.BackendNull}

	var err error

	ctx, err = malgo.InitContext(backends, malgo.ContextConfig{ThreadPriority: malgo.ThreadPriorityRealtime}, func(message string) {
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

	deviceConfig.DeviceType = malgo.Playback
	deviceConfig.Playback.DeviceID = &d.ID
	deviceConfig.Alsa.NoMMap = 0
	deviceConfig.Playback.ShareMode = malgo.Exclusive

	deviceConfig.BufferSizeInMilliseconds = 0
	deviceConfig.BufferSizeInFrames = 0
	deviceConfig.Periods = 10
	deviceConfig.Playback.Channels = uint32(2)
	deviceConfig.SampleRate = uint32(source.SampleRate)
	deviceConfig.Playback.Format = source.SampleFormat

	// This is the function that's used for sending more data to the device for playback.
	onData := func(outputSamples, inputSamples []byte, frameCount uint32) {
		_, err := buffer.Read(outputSamples)
		if err == io.EOF {
			return
		}
	}

	deviceCallbacks := malgo.DeviceCallbacks{
		Data: onData,
	}

	device, err = malgo.InitDevice(ctx.Context, deviceConfig, deviceCallbacks)
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

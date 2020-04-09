package main

import (
	"bytes"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/godsic/malgo"
	"github.com/rivo/tview"
	wav "github.com/youpy/go-wav"
)

var (
	device        = new(malgo.Device)
	deviceConfig  = malgo.DefaultDeviceConfig()
	playerCtl     = make(chan int)
	buffer        bytes.Buffer
	bufferMutex   sync.Mutex
	contextConfig = malgo.ContextConfig{ThreadPriority: malgo.ThreadPriorityRealtime,
		Alsa: malgo.AlsaContextConfig{UseVerboseDeviceEnumeration: 1},
	}
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
		vibeLogger.Println(err)
	}
	ctx.Free()
}

func chooseCard() error {
	backends := []malgo.Backend{malgo.BackendAlsa, malgo.BackendWasapi, malgo.BackendNull}

	var err error

	ctx, err = malgo.InitContext(backends, contextConfig, miniaudioLogger)
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
	deviceConfig.Playback.ShareMode = malgo.Exclusive

	deviceConfig.PeriodSizeInMilliseconds = 4
	deviceConfig.Periods = 16
	deviceConfig.Playback.Channels = uint32(2)
	deviceConfig.SampleRate = uint32(source.dev.(*Source).SampleRate)
	deviceConfig.Playback.Format = source.dev.(*Source).SampleFormat

	deviceConfig.Wasapi.NoAutoConvertSRC = 1
	deviceConfig.Wasapi.NoDefaultQualitySRC = 0
	deviceConfig.Wasapi.NoHardwareOffloading = 1
	deviceConfig.Wasapi.NoAutoStreamRouting = 1

	deviceConfig.Alsa.NoMMap = 0
	deviceConfig.Alsa.NoAutoResample = 1
	deviceConfig.Alsa.NoAutoFormat = 1
	deviceConfig.Alsa.NoAutoChannels = 1

	deviceConfig.NoClip = 1
	deviceConfig.NoPreZeroedOutputBuffer = 0

	// This is the function that's used for sending more data to the device for playback.
	onData := func(outputSamples, inputSamples []byte, frameCount uint32) {
		tIn := time.Now()
		n, _ := buffer.Read(outputSamples)
		tOut := time.Now()
		jd := jitterData{timeIn: tIn, timeOut: tOut, requestedBytes: frameCount, readBytes: n}
		timeChannel <- jd
	}

	deviceCallbacks := malgo.DeviceCallbacks{
		Data: onData,
	}

	device, err = malgo.InitDevice(ctx.Context, deviceConfig, deviceCallbacks)
	if err != nil {
		vibeLogger.Println(err)
		return err
	}
	return nil
}

func play() error {

	err := device.Start()
	if err != nil {
		vibeLogger.Fatal(err)
	}
	return nil
}

func pause() error {
	if device.IsStarted() {
		err := device.Stop()
		if err != nil {
			vibeLogger.Fatal(err)
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

	bufferMutex.Lock()
	buffer.Reset()
	buffer.ReadFrom(r)
	bufferMutex.Unlock()
	return nil
}

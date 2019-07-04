package main

import "github.com/gen2brain/malgo"

type Source struct {
	Name          string
	Vout          float64 // V
	Rl            float64 // Ohm
	SampleRate    uint    // Hz
	SampleBits    int
	SampleFormat  malgo.FormatType
	VolumeControl string
}

func NewSource(name string, Vout, Rl float64, SampleRate uint, SampleBits int, VolumeControl string) *Source {
	var src Source
	src.Name = name
	src.Vout = Vout
	src.Rl = Rl
	src.SampleRate = SampleRate
	src.SampleBits = SampleBits
	src.SampleFormat = bitsPerSampleToDeviceFormat(SampleBits)
	src.VolumeControl = VolumeControl
	return &src
}

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
)

const (
	EMPTY = iota
	SINK
	SOURCE
	CARD
)

type AudioDevice struct {
	dev  interface{}
	Type int
	Name string
}

func NewAudioDevice(which int) *AudioDevice {
	ad := new(AudioDevice)
	ad.Type = which
	return ad
}

func (s *AudioDevice) Set() error {
	switch s.Type {
	case EMPTY:
		return nil
	case SINK:
		for _, v := range Sinks {
			if s.Name == v.Name {
				s.dev = v
				vibeLogger.Printf("Set SINK to %s\n", s.Name)
				return nil
			}
		}
		return errors.New(fmt.Sprintf("Unknown sink: %s", s.Name))
	case SOURCE:
		for _, v := range Sources {
			if s.Name == v.Name {
				s.dev = v
				vibeLogger.Printf("Set SOURCE to %s\n", s.Name)
				return nil
			}
		}
		return errors.New(fmt.Sprintf("Unknown source: %s", s.Name))
	default:
	}
	return errors.New("Unknown type")
}

func (s *AudioDevice) Load(fn string) error {
	t := s.Type
	outBytes, err := ioutil.ReadFile(fn)
	if err != nil {
		return err
	}
	err = json.Unmarshal(outBytes, s)
	if err != nil {
		return err
	}
	if t != s.Type {
		return errors.New("Wrong type")
	}
	return s.Set()
}

func (s *AudioDevice) Save(fn string) error {
	outBytes, err := json.Marshal(s)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(fn, outBytes, 0600)
	return err
}

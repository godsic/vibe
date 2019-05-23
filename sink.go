package main

type Sink struct {
	Name        string
	R           float64
	Sensitivity float64
}

func NewSink(name string, R, Sensitivity float64) *Sink {
	var S Sink
	S.Name = name
	S.R = R
	S.Sensitivity = Sensitivity
	return &S
}

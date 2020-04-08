package main

type Sink struct {
	Name        string
	R           float64
	Sensitivity float64
	SoxArgs     string
}

func NewSink(name string, R, Sensitivity float64, SoxArgs string) *Sink {
	var S Sink
	S.Name = name
	S.R = R
	S.Sensitivity = Sensitivity
	S.SoxArgs = SoxArgs
	return &S
}

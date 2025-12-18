package main

type LogLevel = int

const (
	INFO LogLevel = iota
	ERROR
)

type Device int
type Direction int
type Step int

const (
	DeviceNone Device = iota
	Fluke
	Additel
)

const (
	DirectionNone Direction = iota
	DirectionUp
	DirectionDown
)

const (
	StepNone Step = iota
	Step1
	Step2
)

type Command string

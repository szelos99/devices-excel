package main

type LogLevel = int

const (
	INFO LogLevel = iota
	ERROR
)

const (
	DirectionNone Direction = iota
	DirectionUp
	DirectionDown
)

type Command string

const (
	FlukeIDNCmd Command = "*IDN?"
	FlukeValCmd Command = "VAL?"
	ADTIDNCmd   Command = "255:R:OTYPE:1"
	ADTValCmd   Command = "255:R:MRMD:1"
)

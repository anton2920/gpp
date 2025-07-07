package main

type Target int

const (
	TargetNone = Target(iota)
	TargetSerial
	TargetJSON
	TargetCount
)

var Target2String = [...]string{
	TargetSerial: "serial",
	TargetJSON:   "json",
}

var Target2Generator = [...]Generator{
	TargetSerial: SerialGenerator,
	TargetJSON:   JSONGenerator,
}

package main

type Date int64

//gpp:generate: fill(values), verify
//gpp:verify: InsertBefore={{now := Date(GetServerTime())}}, SOA
type DateRanges struct {
	StartDates []Date //gpp:fill: Func=ParseDate; verify: Max={now}
	EndDates   []Date /*gpp:fill: Func=ParseDate; verify: Min={.StartDate}, Max={now}, Optional*/
}

/*
StartDates = [2026-01-17, 2025-01-16]
EndDates   = [..., 0]
*/

func ParseDate(s string) Date {
	return 0
}

func GetServerTime() int64 {
	return 0
}

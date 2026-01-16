package main

//gpp:generate: fill(values), verify
//gpp:verify: InsertBefore={{now := GetServerTime()}}, SOA
type DateRanges struct {
	StartDates []int64 //gpp:fill: Func=ParseDate; verify: Max={{now}}
	EndDates   []int64 //gpp:fill: Func=ParseDate; verify: Min={{.StartDate}}, Max={{now}}, Optional
}

/*
StartDates = [2026-01-17, 2025-01-16]
EndDates   = [..., 0]
*/

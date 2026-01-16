package main

type Date int64

//gpp:generate: fill(values), verify
//gpp:verify: InsertBefore={{now := GetServerTime()}}, SOA
type DateRanges struct {
	StartDates []Date //gpp:fill: Func={{ParseDate("2006-01-02", ?)}}; verify: Max={{now}}
	EndDates   []Date /*gpp:fill: Func=ParseDate; verify: Min={{.StartDate}}, Max={{now}}, Optional*/
}

//gpp:generate: fill(values)
//gpp:fill: InsertAfter={{.Permissions = GetPermissionsFromValues(vs.GetMany("Permissions"))}}
type Foo struct {
	Permissions Permission //gpp:fill: nop
}

/*
StartDates = [2026-01-17, 2025-01-16]
EndDates   = [..., 0]
*/

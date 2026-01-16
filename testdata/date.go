package main

type Date int64

//gpp:generate: fill(values), verify
//gpp:verify: InsertBefore={{now := GetServerTime()}}, SOA
type DateRanges struct {
	StartDates []Date //gpp:fill: Func={{ParseDate("2006-01-02", ?)}}; verify: Max={{now}}
	EndDates   []Date /*gpp:fill: Func=ParseDate; verify: Min={{.StartDate}}, Max={{now}},
	Func={{VerifyDate(.StartDate, ?)}}, Optional*/
}

//gpp:generate: fill(values)
type Foo struct {
	Permissions Permission //gpp:fill: Func={GetPermissionsFromValues(vs.GetMany("?"))}
}

/*
StartDates = [2026-01-17, 2025-01-16]
EndDates   = [..., 0]
*/

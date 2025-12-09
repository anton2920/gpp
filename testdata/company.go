package main

//gpp:generate: fill, verify
type CompanySearchFilter struct {
	Name    string //gpp:verify: MinLength=1, MaxLength=45, Optional
	HasLogo bool
}

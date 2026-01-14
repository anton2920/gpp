package main

type Permission uint32

//gpp:generate: fill, verify
type CompanyRole struct {
	Name        string     //gpp:verify: MinLength=1, MaxLength=45
	Permissions Permission //gpp:fill: Func={GetCompanyRolePermissionsFromValues(vs.GetMany("?"))}; verify: Required
}

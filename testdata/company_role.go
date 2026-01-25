package main

type Permission uint32

//gpp:generate: fill, verify
//gpp:fill: InsertAfter={{.Permissions = GetCompanyRolePermissionsFromValues(vs.GetMany("Permissions"))}}
type CompanyRole struct {
	Name        string     //gpp:verify: MinLength=1, MaxLength=45
	Permissions Permission //gpp:fill: nop; gpp:verify: Required
}

func GetCompanyRolePermissionsFromValues(xs []string) Permission {
	return 0
}

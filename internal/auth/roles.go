package auth

const (
	RoleAdmin   = "admin"
	RoleScanner = "scanner"
)

func RoleIsValid(role string) bool {
	switch role {
	case RoleAdmin, RoleScanner:
		return true
	default:
		return false
	}
}

// CanAccess returns whether role meets minimum requirement.
func CanAccess(role, required string) bool {
	if required == RoleAdmin {
		return role == RoleAdmin
	}
	// scanner requirement: both admin and scanner
	return role == RoleAdmin || role == RoleScanner
}
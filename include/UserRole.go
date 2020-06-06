package include

// Admin, - see all records
// User, - see only own records
// Viewer - see only record which is assigned to him as viewer

type UserRole struct {
	Model
	User     string
	RoleName string
}

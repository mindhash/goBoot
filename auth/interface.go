package auth

//To Do: import (
//	"github.com/mindhash/goBoot/base"
//)

type Principal interface {
	//Identifier
	Name() string

	// Returns an appropriate HTTPError for unauthorized access -- a 401 if the receiver is
	// the guest user, else 403.
	UnauthError(message string) error

	validate() error

}

// Role is basically the same as Principal, just concrete. Users can inherit channels from Roles.
type Role interface {
	Principal
}

// A User is a Principal that can log in and have multiple Roles.
// A User is a Principal that can log in and have multiple Roles.
type User interface {
	Principal

	// The user's email address.
	Email() string

	// Sets the user's email address.
	SetEmail(string) error

	// If true, the user is unable to authenticate.
	Disabled() bool

	// Sets the disabled property
	SetDisabled(bool)

	// Authenticates the user's password.
	Authenticate(password string) bool

	// Changes the user's password.
	SetPassword(password string)
}
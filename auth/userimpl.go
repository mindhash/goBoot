package auth

import (
	"regexp"
	"net/http"
	"fmt"
	"golang.org/x/crypto/bcrypt" 
	"github.com/mindhash/goBoot/base"
)

const kBcryptCostFactor = bcrypt.DefaultCost  //used for encryption
var kValidEmailRegexp *regexp.Regexp //used for checking email validity


//  implementation of User interface
type userImpl struct {
	
	Name_             string      `json:"name"`
	Email_           string      `json:"email,omitempty"`
	Disabled_        bool        `json:"disabled,omitempty"`
	PasswordHash_    []byte      `json:"passwordhash_bcrypt,omitempty"`
	auth  *Authstore

	//To Do: 
	//	roles []Role
	//	roleImpl // userImpl "inherits from" Role
	//ExplicitRoles_   ch.TimedSet `json:"explicit_roles,omitempty"`
	//RolesSince_      ch.TimedSet `json:"rolesSince"`
}

//To Do: Add Role Impl to drive permission sets
//type roleImpl struct {
//	Name_             string
//add role permissions here
//}


func init() {
	var err error
	kValidEmailRegexp, err = regexp.Compile(`^[-+.\w]+@\w[-.\w]+$`)
	if err != nil {
		panic("Bad kValidEmailRegexp")
	}
}

func IsValidEmail(email string) bool {
	return kValidEmailRegexp.MatchString(email)
}

// Creates a new User object.
func (auth *Authstore) NewUser(username string, password string) (User, error) {
	
	user := &userImpl{
		Name_: username,
		auth: auth,
	}
	
	user.SetPassword(password)
	return user, nil
}

//To Do:
func (auth *Authstore) defaultGuestUser() User {
	user := &userImpl{ Name_:"guest",
	} 
	return user
}

func (user *userImpl) Name() string{
	return user.Name_
}

func (user *userImpl) UnauthError(message string) error {
	if user.Name_ == "" {
		return base.HTTPErrorf(http.StatusUnauthorized, "login required: "+message)
	}
	return base.HTTPErrorf(http.StatusForbidden, message)
}

// Checks whether this userImpl object contains valid data; if not, returns an error.
func (user *userImpl) validate() error {

	if user.Email_ != "" && !IsValidEmail(user.Email_) {
		return base.HTTPErrorf(http.StatusBadRequest, "Invalid email address")
	} 
	
	return nil
}

// Key prefix reserved for user documents in the bucket
const UserKeyPrefix = "_goBoot:user:"


func (user *userImpl) Disabled() bool {
	return user.Disabled_
}

func (user *userImpl) SetDisabled(disabled bool) {
	user.Disabled_ = disabled
}

func (user *userImpl) Email() string {
	return user.Email_
}

func (user *userImpl) SetEmail(email string) error {
	if email != "" && !IsValidEmail(email) {
		return base.HTTPErrorf(http.StatusBadRequest, "Invalid email address")
	}
	user.Email_ = email
	return nil
}

// Returns true if the given password is correct for this user, and the account isn't disabled.
func (user *userImpl) Authenticate(password string) bool {
	if user == nil {
		return false
	} else if user.PasswordHash_ == nil {
		if password != "" {
			return false
		}
	} else if !compareHashAndPassword(user.PasswordHash_, []byte(password)) {
		return false
	}
	return !user.Disabled_
}

// Changes a user's password to the given string.
func (user *userImpl) SetPassword(password string) {
	if password == "" {
		user.PasswordHash_ = nil
	} else {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), kBcryptCostFactor)
		if err != nil {
			panic(fmt.Sprintf("Error hashing password: %v", err))
		}
		user.PasswordHash_ = hash
	}
}



package auth

import (
	"github.com/mindhash/goBoot/base"
	)

//Data store that stores user info
type Authstore struct {
	datastore base.Datastore	 
}

// Creates a new Authenticator that stores user info  
func NewAuthstore(datastore base.Datastore) *Authstore {
	return &Authstore{ 
		datastore,
	}
}


func (auth *Authstore) GetPrincipal(name string, isUser bool) (Principal, error) {
	return auth.GetUser(name)
}

func (auth *Authstore) GetUser(name string) (User, error) {
	princ, err := auth.getPrincipal(name, func() Principal { return &userImpl{} })
	if err != nil {
		return nil, err
	} else if princ == nil {
		if name == "" {
			princ = auth.defaultGuestUser()
		} else {
			return nil, nil
		}
	}
	princ.(*userImpl).auth = auth
	return princ.(User), err
}




func (auth *Authstore) getPrincipal(docID string, factory func() Principal) (Principal, error) {
	return nil, nil
}

// Looks up a User by email address.
func (auth *Authstore) GetUserByEmail(email string) (User, error) {
	return nil, nil
}


// Saves the information for a user/role.
func (auth *Authstore) Save(p Principal) error {
return nil
}

// Deletes a user/role.
func (auth *Authstore) Delete(p Principal) error {
	return nil
}


func (auth *Authstore) AuthenticateUser(username string, password string) User {
	user, _ := auth.GetUser(username)
	if user == nil || !user.Authenticate(password) {
		return nil
	}
	return user
}

func (auth *Authstore) RegisterNewUser(username, email string) (User, error) {
	user, err := auth.NewUser(username, base.GenerateRandomSecret())
	if err != nil {
		return nil, err
	}
	user.SetEmail(email)
	err = auth.Save(user)
	if err != nil {
		return nil, err
	}
	return user, err
}
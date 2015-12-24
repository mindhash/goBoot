package auth

import (
	"github.com/mindhash/goBoot/base"
	)

//Data store that stores user info
type Authstore struct {
	bucket base.Datastore	 
}

// Creates a new Authenticator that stores user info  
func NewAuthstore(bucket base.Datastore) *Authstore {
	return &Authstore{ 
		bucket,
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

func (auth *Authstore) getPrincipal(name string, factory func() Principal) (Principal, error) {
	var princ Principal

	princ := factory()

	err:= auth.bucket.FindByvalue("principals", bson.M{"Name_":name}, princ)

	if (err!=nil) {
		return nil, err	
	}

	return princ, nil
}

// Looks up a User by email address.
func (auth *Authstore) GetUserByEmail(email string) (User, error) {
	
	user := &userImpl{}
	//derive user by email
	err := auth.bucket.FindByValue("principals",bson.M{"Email_": email},user)
	if (err!=nil){
		return nil, err
	}

	return user, nil 
}


// Saves the information for a user/role.
func (auth *Authstore) Save(p Principal) error {

	if err:= p.validate(); err != nil {
		return err
	}

	if user, ok := p.(User); ok {
		//fail if user email already registered 
		if user.Email() != "" {
			userByEmailInfo :=  auth.GetUserByEmail(user.Email())
			if err != nil {
				return err
			}
			if (userByEmailInfo.Name() != user.Name()) {
					//raise error 
				return errors.New("User email already registered")
			}
		}
	}

	err = auth.bucket.Insert ("principals", &p)
	if err != nil {
		return err
	}


	//base.LogTo("Auth", "Saved %s: %s", p._id, data)
	return nil  

}

// Deletes a user/role. To Do
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
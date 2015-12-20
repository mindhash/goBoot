package db

import (
	"net/http"	
	"time"
	"regexp"
	"github.com/mindhash/goBoot/base"
	"github.com/mindhash/goBoot/auth"

)

type Body map[string]interface{}

type DatabaseContext struct {
	Name               string                  // Database name
	Datastore          base.Datastore             // Data Storage
	StartTime          time.Time               // Timestamp when context was instantiated
}


func ConnectToDataStore(spec base.DStoreSpec) (ds base.Datastore, err error){
	ds, err = base.GetDatastore(spec)
	if err != nil {
		err = base.HTTPErrorf(http.StatusBadGateway,
			"Unable to connect to server: %s", err)
	}  
	return
}


func ValidateDatabaseName(dbName string) error {
	// http://wiki.apache.org/couchdb/HTTP_database_API#Naming_and_Addressing
	if match, _ := regexp.MatchString(`^[a-z][-a-z0-9_$()+/]*$`, dbName); !match {
		return base.HTTPErrorf(http.StatusBadRequest,
			"Illegal database name: %s", dbName)
	}
	return nil
}

// Creates a new DatabaseContext on a bucket. The bucket will be closed when this context closes.
func NewDatabaseContext(dbName string, ds base.Datastore) (*DatabaseContext, error) {
	if err := ValidateDatabaseName(dbName); err != nil {
		return nil, err
	}
	context := &DatabaseContext{
		Name:       dbName,
		Datastore:     ds,
		StartTime:  time.Now(),
	}
		return context, nil
}
 

//close DB Session
func (context *DatabaseContext) Close() { 
	context.Datastore.Close()
	context.Datastore = nil
}

func (context *DatabaseContext) IsClosed() bool {
	return context.Datastore == nil
}

func (context *DatabaseContext) Authstore() *auth.Authstore {
	return auth.NewAuthstore(context.Datastore)
}
 



 
  
package rest

import (
	"net/http"
	"sync"
	"github.com/mindhash/goBoot/db" 
	"github.com/mindhash/goBoot/base"
)

type ServerContext struct {
	config      *ServerConfig
	database_  *db.DatabaseContext 
	lock        sync.RWMutex
//	statsTicker *time.Ticker
	HTTPClient  *http.Client
}

func NewServerContext(config *ServerConfig) *ServerContext {
	sc := &ServerContext{
		config:     config, 
		HTTPClient: http.DefaultClient,
	}
	return sc
}

func (sc *ServerContext) GetDatabase() (db *db.DatabaseContext, err error) {
	return sc.database_, nil
}

func (sc *ServerContext) AddDatabaseFromConfig(dbc *DbConfig) (db *db.DatabaseContext, err error){
	return sc.GetOrAddDatabase(dbc)
}

// get or add database context to server context
func (sc *ServerContext) GetOrAddDatabase(dbc *DbConfig) (context *db.DatabaseContext, err error) {
	
	sc.lock.Lock()
	defer sc.lock.Unlock()

	if (sc.database_ !=nil) {
		return sc.database_,nil
	}
  
	//get data store 
	ds, err := db.ConnectToDataStore (base.DStoreSpec{Hostaddr: dbc.Host , Dbname: dbc.Name, Dbuser: dbc.User, Dbpwd: dbc.Password})


	if err!=nil {
		return nil,err
	}

	context,err= db.NewDatabaseContext(dbc.Name, ds)

	return 

}

func (sc *ServerContext) CloseDatabase() {
	sc.database_.Close()
}
package rest

import (
	"net/http"
	"sync" 
	//"github.com/mindhash/goBackend/base"
)

type ServerContext struct {
	config      *ServerConfig
	//database_  *db.DatabaseContext// databases_ map[string]*db.DatabaseContext
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

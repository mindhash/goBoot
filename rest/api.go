package rest

import (
	"fmt"
)

const ServerName = "goBackend Server"
const VersionNumber float64 = 1.0                    // API/feature level


var LongVersionString string

func init(){
	LongVersionString = fmt.Sprintf("%s/unofficial", ServerName)
}

// HTTP handler for the root ("/")
func (h *handler) handleRoot() error {
	response := map[string]interface{}{
		"goBackend": "Welcome",
		"version": LongVersionString,
		"vendor":  map[string]interface{}{"name": ServerName, "version": VersionNumber},
	}
	
	h.writeJSON(response)
	return nil
}
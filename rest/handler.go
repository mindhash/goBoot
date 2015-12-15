package rest


import (
	"bytes"
	"encoding/json"
	"expvar"
	"fmt"
	"io"
	//"io/ioutil"
	"mime"
	//"mime/multipart"
	"net/http"
	//"net/url"
	//"os"
	//"strconv"
	"strings"
	"sync/atomic"
	"time"
	//"github.com/gorilla/mux"
 
	"github.com/mindhash/goBackend/base"
	"github.com/mindhash/goBackend/db"
	//"github.com/mindhash/goBackend/auth"
)

var kNotFoundError = base.HTTPErrorf(http.StatusNotFound, "missing")
var kBadMethodError = base.HTTPErrorf(http.StatusMethodNotAllowed, "Method Not Allowed")
var kBadRequestError = base.HTTPErrorf(http.StatusMethodNotAllowed, "Bad Request")

// If set to true, JSON output will be pretty-printed.
var PrettyPrint bool = false

var restExpvars = expvar.NewMap("gobackend_rest")

var lastSerialNum uint64 = 0

// Encapsulates the state of handling an HTTP request.
type handler struct {
	server         *ServerContext
	rq             *http.Request
	response       http.ResponseWriter
	status         int
	statusMessage  string
	requestBody    io.ReadCloser
	//To Do:db             *db.Database
	//To Do:user           auth.User
	privs          handlerPrivs
	startTime      time.Time
	serialNumber   uint64
	loggedDuration bool
}

type handlerPrivs int


type handlerMethod func(*handler) error


const (
	regularPrivs = iota // Handler requires authentication
	publicPrivs         // Handler checks auth but doesn't require it
	adminPrivs          // Handler ignores auth, always runs with root/admin privs
)


// Creates an http.Handler that will run a handler with the given method
func makeHandler(server *ServerContext, privs handlerPrivs, method handlerMethod) http.Handler {
	return http.HandlerFunc(func(r http.ResponseWriter, rq *http.Request) {
		h := newHandler(server, privs, r, rq)
		err := h.invoke(method)
		h.writeError(err)
		h.logDuration(true) 
	})
}

func newHandler(server *ServerContext, privs handlerPrivs, r http.ResponseWriter, rq *http.Request) *handler {
	return &handler{
		server:       server,
		privs:        privs,
		rq:           rq,
		response:     r,
		status:       http.StatusOK,
		serialNumber: atomic.AddUint64(&lastSerialNum, 1),
		startTime:    time.Now(),
	}
}

// Top-level handler call. It's passed a pointer to the specific method to run.
func (h *handler) invoke(method handlerMethod) error {
	// exp vars used for reading request counts
	restExpvars.Add("requests_total", 1)
	restExpvars.Add("requests_active", 1)
	defer restExpvars.Add("requests_active", -1)

	switch h.rq.Header.Get("Content-Encoding") {
	case "":
		h.requestBody = h.rq.Body
	default:
		return base.HTTPErrorf(http.StatusUnsupportedMediaType, "Unsupported Content-Encoding;")
	}
	
	
	h.logRequestLine()

	//assign db to handler h

	return method(h) // Call the actual handler code
	
}

func (h *handler) logRequestLine() {
	if !base.LogKeys["HTTP"] {
		return
	}
	as := ""
	//To Do: if h.privs == adminPrivs {
	//	as = "  (ADMIN)"
	//} else if h.user != nil && h.user.Name() != "" {
	//	as = fmt.Sprintf("  (as %s)", h.user.Name())
	//}
	base.LogTo("HTTP", " #%03d: %s %s%s", h.serialNumber, h.rq.Method, h.rq.URL, as)
}

func (h *handler) logDuration(realTime bool) {
	if h.loggedDuration {
		return
	}
	h.loggedDuration = true

	var duration time.Duration
	if realTime {
		duration = time.Since(h.startTime)
		bin := int(duration/(100*time.Millisecond)) * 100
		restExpvars.Add(fmt.Sprintf("requests_%04dms", bin), 1)
	}

	logKey := "HTTP+"
	if h.status >= 300 {
		logKey = "HTTP"
	}
	base.LogTo(logKey, "#%03d:     --> %d %s  (%.1f ms)",
		h.serialNumber, h.status, h.statusMessage,
		float64(duration)/float64(time.Millisecond))
}

// Used for indefinitely-long handlers like _changes that we don't want to track duration of
func (h *handler) logStatus(status int, message string) {
	h.setStatus(status, message)
	h.logDuration(false) // don't track actual time
}


func (h *handler) userAgentIs(agent string) bool {
	userAgent := h.rq.Header.Get("User-Agent")
	return len(userAgent) > len(agent) && userAgent[len(agent)] == '/' && strings.HasPrefix(userAgent, agent)
}

func (h *handler) requestAccepts(mimetype string) bool {
	accept := h.rq.Header.Get("Accept")
	return accept == "" || strings.Contains(accept, mimetype) || strings.Contains(accept, "*/*")
}

func (h *handler) setHeader(name string, value string) {
	h.response.Header().Set(name, value)
}

func (h *handler) setStatus(status int, message string) {
	h.status = status
	h.statusMessage = message
}

func (h *handler) flush() {
	switch r := h.response.(type) {
	case http.Flusher:
		r.Flush()
	}
}


func (h *handler) writeError(err error) {
	if err != nil {
		status, message := base.ErrorAsHTTPStatus(err)
		h.writeStatus(status, message)
	}
}

// Writes the response status code, and if it's an error writes a JSON description to the body.
func (h *handler) writeStatus(status int, message string) {
	if status < 300 {
		h.response.WriteHeader(status)
		h.setStatus(status, message)
		return
	}
	// Got an error:
	var errorStr string
	switch status {
	case http.StatusNotFound:
		errorStr = "not_found"
	case http.StatusConflict:
		errorStr = "conflict"
	default:
		errorStr = http.StatusText(status)
		if errorStr == "" {
			errorStr = fmt.Sprintf("%d", status)
		}
	}

//	h.disableResponseCompression()
	h.setHeader("Content-Type", "application/json")
	h.response.WriteHeader(status)
	h.setStatus(status, message)
	jsonOut, _ := json.Marshal(db.Body{"error": errorStr, "reason": message})
	h.response.Write(jsonOut)
}



// Writes an object to the response in JSON format.
// If status is nonzero, the header will be written with that status.
func (h *handler) writeJSONStatus(status int, value interface{}) {
	if !h.requestAccepts("application/json") {
		base.Warn("Client won't accept JSON, only %s", h.rq.Header.Get("Accept"))
		h.writeStatus(http.StatusNotAcceptable, "only application/json available")
		return
	}
	 
	jsonOut, err := json.Marshal(value)
	 
	if err != nil {
		base.Warn("Couldn't serialize JSON for %v : %s", value, err)
		h.writeStatus(http.StatusInternalServerError, "JSON serialization failed")
		return
	}

	if PrettyPrint {
		var buffer bytes.Buffer
		json.Indent(&buffer, jsonOut, "", "  ")
		jsonOut = append(buffer.Bytes(), '\n')
	}
		
	h.setHeader("Content-Type", "application/json")
	if h.rq.Method != "HEAD" {
		//if len(jsonOut) < 1000 {
		//	h.disableResponseCompression()
		//}
		h.setHeader("Content-Length", fmt.Sprintf("%d", len(jsonOut)))
		if status > 0 {
			h.response.WriteHeader(status)
			h.setStatus(status, "")
		}
		h.response.Write(jsonOut)
	} else if status > 0 {
		h.response.WriteHeader(status)
		h.setStatus(status, "")
	}
}

func (h *handler) writeJSON(value interface{}) {
	h.writeJSONStatus(http.StatusOK, value)
}



// Parses a JSON request body, returning it as a Body map.
func (h *handler) readJSON() (db.Body, error) {
	var body db.Body
	return body, h.readJSONInto(&body)
}

// Parses a JSON request body into a custom structure.
func (h *handler) readJSONInto(into interface{}) error {
	
	contentType := h.rq.Header.Get("Content-Type")
	if contentType != "" && !strings.HasPrefix(contentType, "application/json") {
		return base.HTTPErrorf(http.StatusUnsupportedMediaType, "Invalid content type %s", contentType)
	}
 
 	//TO DO: zip version to be added
	  	
	decoder := json.NewDecoder(h.requestBody)
	if err := decoder.Decode(into); err != nil {
		base.Warn("Couldn't parse JSON in HTTP request: %v", err)
		return base.HTTPErrorf(http.StatusBadRequest, "Bad JSON")
	}
	 
	return nil
}

//TO DO: Need to add multi part reads
//This function handles marshaling of input JSON into Struct
func (h *handler) readObject(obj interface{}) ( interface{}, error){
	
	contentType, _ , _ := mime.ParseMediaType(h.rq.Header.Get("Content-Type"))
	
	//process JSON Documents only
	switch contentType {
		
	case "", "application/json":
		return obj, h.readJSONInto(obj)
	default:
		return nil, base.HTTPErrorf(http.StatusUnsupportedMediaType, "Invalid content type %s", contentType)
	}
}

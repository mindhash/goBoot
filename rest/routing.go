package rest

import (
	"github.com/gorilla/mux"
	"net/http"  
	"strings"
	"regexp"
)

// Regex that matches a URI containing a regular doc ID with an escaped "/" character
var docWithSlashPathRegex *regexp.Regexp

func init() {
	docWithSlashPathRegex, _ = regexp.Compile("/"  + "/[^_].*%2[fF]")
}

func createHandler(sc *ServerContext, privs handlerPrivs) (*mux.Router) {

	r := mux.NewRouter()
	r.StrictSlash(true)
	
	// Global operations:
	r.Handle("/", makeHandler(sc, privs, (*handler).handleRoot)).Methods("GET", "HEAD")
		
	return r
}


// Creates the HTTP handler for the public API 
func CreatePublicHandler(sc *ServerContext) http.Handler {

	r := createHandler(sc, regularPrivs)
	return wrapRouter(sc, regularPrivs, r)
}


// Returns a top-level HTTP handler for a Router. Calls specific functions from handler
func wrapRouter(sc *ServerContext, privs handlerPrivs, router *mux.Router) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, rq *http.Request) {
		fixQuotedSlashes(rq)
		var match mux.RouteMatch

		if router.Match(rq, &match) {
			router.ServeHTTP(response, rq)
		} else {
			// Log the request
			h := newHandler(sc, privs, response, rq)
			h.logRequestLine()

			// What methods would have matched?
			var options []string
			for _, method := range []string{"GET", "HEAD", "POST", "PUT", "DELETE"} {
				if wouldMatch(router, rq, method) {
					options = append(options, method)
				}
			}
			if len(options) == 0 {
				h.writeStatus(http.StatusNotFound, "unknown URL")
			} else {
				response.Header().Add("Allow", strings.Join(options, ", "))

				if rq.Method != "OPTIONS" {
					h.writeStatus(http.StatusMethodNotAllowed, "")
				} else {
					h.writeStatus(http.StatusNoContent, "")
				}
			}
			h.logDuration(true)
		}
	})
}


func matchedOrigin(allowOrigins []string, rqOrigins []string) string {
	for _, rv := range rqOrigins {
		for _, av := range allowOrigins {
			if rv == av {
				return av
			}
		}
	}
	for _, av := range allowOrigins {
		if av == "*" {
			return "*"
		}
	}
	return ""
}

func fixQuotedSlashes(rq *http.Request) {
	uri := rq.RequestURI
	if docWithSlashPathRegex.MatchString(uri) {
		if stop := strings.IndexAny(uri, "?#"); stop >= 0 {
			uri = uri[0:stop]
		}
		rq.URL.Path = uri
	}
}

func wouldMatch(router *mux.Router, rq *http.Request, method string) bool {
	savedMethod := rq.Method
	rq.Method = method
	defer func() { rq.Method = savedMethod }()
	var matchInfo mux.RouteMatch
	return router.Match(rq, &matchInfo)
}
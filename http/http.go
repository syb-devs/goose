package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/dimfeld/httptreemux"
	"github.com/syb-devs/goose"
)

// ErrForbidden represents an HTTP 403 error
var ErrForbidden = NewError(403, "Forbidden")

// HandlerFunc is used for HTTP handlers. In addition to the ResponseWriter and Request objects, they receive a Context object
// and should return an error
type HandlerFunc func(http.ResponseWriter, *http.Request, *Context) error

// Error represents an HTTP Error
type Error struct {
	Code    int
	Message string
}

// NewError returns a new Error with the given HTTP code and message
func NewError(code int, message string) *Error {
	if message == "" {
		message = http.StatusText(code)
	}
	return &Error{Code: code, Message: message}
}

// Error returns the HTTP Error message as a string
func (e Error) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.Code, e.Message)
}

// URLParams contains a map with the parameter names and values from the URL as defined in the routes
type URLParams map[string]string

func (ps URLParams) ByName(name string) string {
	return ps[name]
}

// HandlerAdapter is a function adapter which converts a Goose HandlerFunc to a httptreemux.HandlerFunc
func HandlerAdapter(f HandlerFunc) httptreemux.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, ps map[string]string) {
		goose.Log.Debug(fmt.Sprintf("%v %v", r.Method, r.URL.Path))

		defer panicHandler(w, r)

		ctx, err := newContext(r, ps)
		defer ctx.close()

		if err != nil {
			panic(fmt.Sprintf("error creating a new context: %v", err))
		}

		// Run the wrapped handler
		err = f(w, r, ctx)
		if err != nil {
			handleError(w, r, err)
			return
		}
	}
}

func panicHandler(w http.ResponseWriter, r *http.Request) {
	if p := recover(); p != nil {
		goose.Log.Critical(fmt.Sprintf("recovering from panic: %v. \nstack trace: %s", p, debug.Stack()))

		if err, ok := p.(error); ok {
			handleError(w, r, err)
		}
	}
}

func handleError(w http.ResponseWriter, r *http.Request, err error) {
	goose.Log.Error(fmt.Sprintf("error: %T: %v", err, err))

	var respErr *Error
	if httpErr, ok := err.(*Error); ok {
		respErr = httpErr
	} else {
		respErr = NewError(500, "")
	}
	WriteJSON(w, respErr.Code, respErr)
}

func enableCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			h.ServeHTTP(w, r)
			return
		}
		// Enable CORS for the origin
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// If it's a preflight request, just return
		if r.Method == "OPTIONS" {
			return
		}

		// Once CORS headers are set, forward the request in the chain
		h.ServeHTTP(w, r)
	})
}

// WriteJSON sets the Content-Type header of the given ResponseWriter to JSON an writes the given data to it
func WriteJSON(w http.ResponseWriter, status int, d interface{}) error {
	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", "application/json")
	if status > 0 {
		w.WriteHeader(status)
	}
	return enc.Encode(d)
}

// ProcessError can convert between errors. For example, it can map model-layer errors (like a not found from database)
// to HTTP-layer errors (the equivalent 404 Not Found error)
func ProcessError(err error) error {
	if err == goose.ErrNotFound {
		return NewError(404, "Not found")
	}
	if err == goose.ErrInvalidIDFormat {
		return NewError(400, err.Error())
	}
	return err
}

// GetSubdomain returns the subdomain string from the given Request object
func GetSubdomain(r *http.Request) string {
	i := strings.Index(r.Host, ".")
	return r.Host[0:i]
}

// PrefixSlash adds a slash "/" at the beginning of the given string if not already present
func PrefixSlash(path string) string {
	if path[0:1] == "/" {
		return path
	}
	return "/" + path
}

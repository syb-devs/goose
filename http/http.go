package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strings"

	"bitbucket.org/syb-devs/goose"
	"github.com/dimfeld/httptreemux"
)

var ErrForbidden = NewError(403, "Forbidden")

type HandlerFunc func(http.ResponseWriter, *http.Request, *Context) error

type Error struct {
	Code    int
	Message string
}

func NewError(code int, message string) *Error {
	return &Error{Code: code, Message: message}
}

func (e Error) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.Code, e.Message)
}

type URLParams map[string]string

func (ps URLParams) ByName(name string) string {
	return ps[name]
}

// HandlerAdapter is a function adapter which converts a Goose HandlerFunc to a httptreemux.HandlerFunc
func HandlerAdapter(f HandlerFunc) httptreemux.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, ps map[string]string) {
		log.Printf("\nserving %v", r.URL.RequestURI())
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
		log.Printf("recovering from panic: %v %s", p, debug.Stack())

		if err, ok := p.(error); ok {
			handleError(w, r, err)
		}
	}
}

func handleError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("error: %T: %v", err, err)

	if httpErr, ok := err.(*Error); ok {
		http.Error(w, httpErr.Error(), httpErr.Code)
	} else {
		http.Error(w, "Server error", 500)
	}
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

func WriteJSON(w http.ResponseWriter, d interface{}) error {
	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", "application/json")
	return enc.Encode(d)
}

func ProcessError(err error) error {
	if err == goose.ErrNotFound {
		return NewError(404, "Not found")
	}
	if err == goose.ErrInvalidIDFormat {
		return NewError(400, err.Error())
	}
	return err
}

func GetSubdomain(r *http.Request) string {
	i := strings.Index(r.Host, ".")
	return r.Host[0:i]
}

func PrefixSlash(path string) string {
	if path[0:1] == "/" {
		return path
	}
	return "/" + path
}

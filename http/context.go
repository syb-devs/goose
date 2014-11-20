package http

import (
	"log"
	"net/http"

	"bitbucket.org/syb-devs/goose"
)

// context holds data which should be isolated for a single request
type Context struct {
	DB        *goose.DBConn
	URLParams URLParams
	User      *goose.User
}

// NewContext returns a context object to be used in the current request, with
// a copy of the database connection
func newContext(r *http.Request, ps map[string]string) (*Context, error) {
	log.Printf("%v", ps)
	return &Context{
		DB:        goose.DefaultDBConn().Copy(),
		URLParams: URLParams(ps),
		User:      &goose.User{}, //TODO: retrieve user from JWT auth header
	}, nil
}

// close closes the context, freeing resources like database connection
// and persisting the session
func (ctx *Context) close() error {
	if ctx == nil {
		return nil
	}

	if ctx.DB != nil {
		ctx.DB.Close()
	}
	return nil
}

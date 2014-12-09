package goose

import (
	"errors"
	"log"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// DBConn represents a database connection, and wraps the MGO Database object
type DBConn struct {
	*mgo.Database
}

var (
	// ErrNotFound error is returned by object repositories when the requested object can not be found
	ErrNotFound = mgo.ErrNotFound
	// ErrInvalidIDFormat error is returned when an invalid ObjectID is given
	ErrInvalidIDFormat = errors.New("invalid format for resource ID")
)

// DBOptions struct contains settings to create a new DBConn
type DBOptions struct {
	URL          string
	Database     string
	SetAsDefault bool
	Debug        bool
}

// DBInitTask is a function that receives a DB connection and is called when a new connection is set up
type DBInitTask func(db *DBConn) error

var defaultDBConn *DBConn

var dbInitTaks = make([]DBInitTask, 0)

// RegisterDBInitTask registers a new DBInitTask that will be run when a new database connection is created
func RegisterDBInitTask(task DBInitTask) {
	dbInitTaks = append(dbInitTaks, task)
}

// NewDBConn creates a new DB connection with the given options
func NewDBConn(ops DBOptions) *DBConn {
	if ops.Debug {
		mgo.SetLogger(&dbLogger{})
		mgo.SetDebug(true)
	}
	session, err := mgo.Dial(ops.URL)
	if err != nil {
		panic(err)
	}
	session.SetMode(mgo.Monotonic, true)
	db := &DBConn{Database: session.DB(ops.Database)}
	if ops.SetAsDefault {
		SetDefaultDBConn(db)
	}
	for _, initTask := range dbInitTaks {
		err := initTask(db)
		if err != nil {
			panic(err)
		}
	}
	return db
}

// Close closes the DB connection's session with the server
func (c *DBConn) Close() {
	if c == nil {
		return
	}
	c.Database.Session.Close()
}

// Copy creates a copy of the DB connection. IMPORTANT: close the copied connection when no longer needed
func (c *DBConn) Copy() *DBConn {
	if c == nil {
		return nil
	}

	sess := c.Session.Copy()
	return &DBConn{
		Database: sess.DB(c.Database.Name),
	}
}

// SetDefaultDBConn sets the globally accessible DB connection
func SetDefaultDBConn(db *DBConn) {
	defaultDBConn = db
}

// DefaultDBConn returns the global DB connection
func DefaultDBConn() *DBConn {
	return defaultDBConn
}

type dbLogger struct{}

// Output is called from the Mgo driver to log debug messages when debug mode is enabled
func (l *dbLogger) Output(calldepth int, s string) error {
	log.Println(s)
	return nil
}

func ValidObjectID(ID string) bool {
	return bson.IsObjectIdHex(ID)
}

func checkObjectId(ID string) error {
	if !ValidObjectID(ID) {
		return ErrInvalidIDFormat
	}
	return nil
}

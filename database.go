package goose

import (
	"errors"
	"log"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type DBConn struct {
	*mgo.Database
}

var (
	ErrNotFound        = mgo.ErrNotFound
	ErrInvalidIDFormat = errors.New("invalid format for resource ID")
)

type DBOptions struct {
	URL          string
	Database     string
	SetAsDefault bool
	Debug        bool
}

type DBInitTask func(db *DBConn) error

var defaultDBConn *DBConn

var dbInitTaks = make([]DBInitTask, 0)

func RegisterDBInitTask(task DBInitTask) {
	dbInitTaks = append(dbInitTaks, task)
}

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

func (c *DBConn) Close() {
	if c == nil {
		return
	}
	c.Database.Session.Close()
}

func (c *DBConn) Copy() *DBConn {
	if c == nil {
		return nil
	}

	sess := c.Session.Copy()
	return &DBConn{
		Database: sess.DB(c.Database.Name),
	}
}

func SetDefaultDBConn(db *DBConn) {
	defaultDBConn = db
}

func DefaultDBConn() *DBConn {
	return defaultDBConn
}

type dbLogger struct{}

func (l *dbLogger) Output(calldepth int, s string) error {
	log.Println(s)
	return nil
}

func checkObjectId(ID string) error {
	if !bson.IsObjectIdHex(ID) {
		return ErrInvalidIDFormat
	}
	return nil
}

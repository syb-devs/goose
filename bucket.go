package goose

import (
	"bitbucket.org/syb-devs/gotools/time"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func init() {
	RegisterDBInitTask(func(db *DBConn) error { return NewBucketRepo(db).Init() })
}

type Bucket struct {
	ID          bson.ObjectId `bson:"_id" json:"id"`
	Name        string        `bson:"name" json:"name"`
	Collection  string        `json:"collection"`
	Objects     int           `json:"collection"`
	Size        int           `json:"collection"`
	time.Stamps `bson:",inline"`
}

type bucketRepo struct {
	col        *mgo.Collection
	db         *DBConn
	privileged bool
	language   string
}

func NewBucketRepo(db *DBConn) *bucketRepo {
	return &bucketRepo{db: db, col: db.C("buckets")}
}

func (fr *bucketRepo) Init() error {
	index := mgo.Index{
		Key:        []string{"name"},
		Unique:     true,
		DropDups:   false,
		Background: false,
		Sparse:     false,
	}
	return fr.col.EnsureIndex(index)
}

func (r *bucketRepo) Insert(c *Bucket) error {
	if c.ID.Hex() == "" {
		c.ID = bson.NewObjectId()
	}
	c.Touch()
	return r.col.Insert(c)
}

func (r *bucketRepo) FindId(ID string) (*Bucket, error) {
	b := &Bucket{}
	if err := checkObjectId(ID); err != nil {
		return b, err
	}
	err := r.col.FindId(bson.ObjectIdHex(ID)).One(b)
	return b, err
}

func (r *bucketRepo) Update(b *Bucket) error {
	b.Touch()
	return r.col.UpdateId(b.ID, b)
}

func (r *bucketRepo) DeleteId(ID string) error {
	if err := checkObjectId(ID); err != nil {
		return err
	}
	return r.col.RemoveId(bson.ObjectIdHex(ID))
}

func (r *bucketRepo) FindName(name string) (*Bucket, error) {
	b := &Bucket{}
	err := r.col.Find(bson.M{"name": name}).One(b)
	return b, err
}

func (r *bucketRepo) Exists(name string) bool {
	_, err := r.FindName(name)
	return err == nil
}

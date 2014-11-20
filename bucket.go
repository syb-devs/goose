package goose

import (
	"bitbucket.org/syb-devs/gotools/time"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Bucket struct {
	ID          bson.ObjectId `bson:"_id" json:"Id"`
	Name        string
	Collection  string
	Objects     int
	Size        int
	time.Stamps `bson:",inline" `
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

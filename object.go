package goose

import (
	"io"
	"log"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func init() {
	RegisterDBInitTask(func(db *DBConn) error { return NewObjectRepo(db).Init() })
}

type Object struct {
	*mgo.GridFile
	metadata *ObjectMetadata
}

func (f *Object) Id() bson.ObjectId {
	return f.GridFile.Id().(bson.ObjectId)
}

func (f *Object) Metadata() *ObjectMetadata {
	if f.metadata == nil {
		f.metadata = &ObjectMetadata{}
		f.GetMeta(&f.metadata)
	}
	return f.metadata
}

type ObjectList struct {
	objects []*Object
	iter    *mgo.Iter
}

func (fl *ObjectList) Objects() []*Object {
	return fl.objects
}

func (fl *ObjectList) Close() error {
	if fl == nil || fl.iter == nil {
		return nil
	}
	return fl.iter.Close()
}

type ObjectMetadata struct {
	BucketID    bson.ObjectId          `bson:"bucketId"`
	UploaderID  bson.ObjectId          `bson:"uploaderId,omitempty"`
	Title       string                 `bson:"title,omitempty"`
	Description string                 `bson:"description,omitempty"`
	Tags        []string               `bson:"tags,omitempty"`
	Custom      map[string]interface{} `bson:"custom,omitempty"`
}

type objectRepo struct {
	gfs *mgo.GridFS
	db  *DBConn
}

func NewObjectRepo(db *DBConn) *objectRepo {
	return &objectRepo{gfs: db.GridFS("fs"), db: db}
}

func (or *objectRepo) Init() error {
	index := mgo.Index{
		Key:        []string{"filename", "metadata.bucketId"},
		Unique:     false,
		Background: false,
		Sparse:     false,
	}
	return or.gfs.Files.EnsureIndex(index)
}

func (or *objectRepo) create(name, ctype string, metadata *ObjectMetadata) (*Object, error) {
	gObject, err := or.gfs.Create(name)
	if err != nil {
		return nil, err
	}
	object := &Object{
		GridFile: gObject,
	}
	object.SetName(name)
	if ctype != "" {
		object.SetContentType(ctype)
	}
	if metadata != nil {
		object.SetMeta(metadata)
	}
	return object, nil
}

func (or *objectRepo) Create(r io.Reader, name, ctype string, metadata *ObjectMetadata) (*Object, error) {
	object, err := or.create(name, ctype, metadata)
	if err != nil {
		return object, err
	}
	defer object.Close()

	_, err = io.Copy(object, r)
	return object, err
}

func (r *objectRepo) UpdateMetada(ID, name string, metadata ObjectMetadata) error {
	log.Printf("ObjectID: %v", ID)
	return r.gfs.Files.Update(bson.M{"_id": bson.ObjectIdHex(ID)}, bson.M{"$set": bson.M{"metadata": metadata, "objectname": name}})
}

func (r *objectRepo) OpenId(id string) (*Object, error) {
	gridFile, err := r.gfs.OpenId(bson.ObjectIdHex(id))
	if err != nil {
		return nil, err
	}
	return &Object{GridFile: gridFile}, nil
}

func (r *objectRepo) Open(filename string, ID bson.ObjectId) (*Object, error) {
	gridFile, err := r.gfs.Open(filename)
	if err != nil {
		return nil, err
	}
	return &Object{GridFile: gridFile}, nil
}

func (r *objectRepo) DeleteId(id string) error {
	return r.gfs.RemoveId(bson.ObjectIdHex(id))
}

func (r *objectRepo) iterToObject(iter *mgo.Iter) (*Object, error) {
	var f *mgo.GridFile
	if r.gfs.OpenNext(iter, &f) {
		return &Object{GridFile: f}, nil
	}
	return nil, mgo.ErrNotFound
}

func (r *objectRepo) iterToObjectList(iter *mgo.Iter) *ObjectList {
	fl := ObjectList{iter: iter}

	var f *mgo.GridFile
	for r.gfs.OpenNext(iter, &f) {
		fl.objects = append(fl.objects, &Object{GridFile: f})
	}
	return &fl
}

func (r *objectRepo) All() (*ObjectList, error) {
	where := bson.M{}
	iter := r.gfs.Find(where).Sort("-uploadDate").Iter()
	return r.iterToObjectList(iter), nil
}

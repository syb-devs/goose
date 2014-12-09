package goose

import (
	"io"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func init() {
	RegisterDBInitTask(func(db *DBConn) error { return NewObjectRepo(db).Init() })
}

type Object struct {
	ID          bson.ObjectId   `json:"id"`
	UploadDate  time.Time       `json:"uploadDate"`
	Size        int64           `json:"size"`
	MD5         string          `json:"md5"`
	Name        string          `json:"name"`
	ContentType string          `json:"contentType"`
	Metadata    *ObjectMetadata `json:"metadata"`
	gf          *mgo.GridFile
}

func newObjectFromGridFile(f *mgo.GridFile, fill bool) *Object {
	obj := &Object{gf: f}
	if fill {
		obj.FillData()
	}
	return obj
}

func (o *Object) GetID() bson.ObjectId {
	return o.gf.Id().(bson.ObjectId)
}

func (o *Object) GridFile() *mgo.GridFile {
	return o.gf
}

func (o *Object) FillData() {
	o.ID = o.gf.Id().(bson.ObjectId)
	o.UploadDate = o.gf.UploadDate()
	o.Size = o.gf.Size()
	o.MD5 = o.gf.MD5()
	o.Name = o.gf.Name()
	o.ContentType = o.gf.ContentType()
	if o.Metadata == nil {
		o.Metadata = &ObjectMetadata{}
		o.gf.GetMeta(&o.Metadata)
	}
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
	BucketID    bson.ObjectId          `bson:"bucketId" json:"bucketId,omitempty"`
	UploaderID  bson.ObjectId          `bson:"uploaderId,omitempty" json:"uploaderId,omitempty"`
	Title       string                 `bson:"title,omitempty" json:"title"`
	Description string                 `bson:"description,omitempty" json:"description"`
	Tags        []string               `bson:"tags,omitempty" json:"tags"`
	Custom      map[string]interface{} `bson:"custom,omitempty" json:"custom"`
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
		gf: gObject,
	}
	object.gf.SetName(name)
	if ctype != "" {
		object.gf.SetContentType(ctype)
	}
	if metadata != nil {
		object.gf.SetMeta(metadata)
	}
	return object, nil
}

func (or *objectRepo) Create(r io.Reader, name, ctype string, metadata *ObjectMetadata) (*Object, error) {
	object, err := or.create(name, ctype, metadata)
	if err != nil {
		return object, err
	}
	defer object.gf.Close()

	_, err = io.Copy(object.gf, r)
	object.FillData()
	return object, err
}

func (r *objectRepo) UpdateMetada(ID, name string, metadata ObjectMetadata) error {
	return r.gfs.Files.Update(bson.M{"_id": bson.ObjectIdHex(ID)}, bson.M{"$set": bson.M{"metadata": metadata, "objectname": name}})
}

func (r *objectRepo) OpenId(ID string) (*Object, error) {
	gridFile, err := r.gfs.OpenId(bson.ObjectIdHex(ID))
	if err != nil {
		return nil, err
	}
	return newObjectFromGridFile(gridFile, true), nil
}

func (r *objectRepo) Open(filename string) (*Object, error) {
	gridFile, err := r.gfs.Open(filename)
	if err != nil {
		return nil, err
	}
	return newObjectFromGridFile(gridFile, true), nil
}

func (r *objectRepo) OpenFromBucket(filename string, bucketID bson.ObjectId) (*Object, error) {
	iter := r.gfs.Find(bson.M{"filename": filename, "metadata.bucketId": bucketID}).Sort("-uploadDate").Iter()
	return r.iterToObject(iter)
}

func (r *objectRepo) DeleteId(ID string) error {
	return r.gfs.RemoveId(bson.ObjectIdHex(ID))
}

func (r *objectRepo) iterToObject(iter *mgo.Iter) (*Object, error) {
	var f *mgo.GridFile
	if r.gfs.OpenNext(iter, &f) {
		return newObjectFromGridFile(f, true), nil
	}
	return nil, mgo.ErrNotFound
}

func (r *objectRepo) iterToObjectList(iter *mgo.Iter) *ObjectList {
	fl := ObjectList{iter: iter}

	var f *mgo.GridFile
	for r.gfs.OpenNext(iter, &f) {
		fl.objects = append(fl.objects, newObjectFromGridFile(f, true))
	}
	return &fl
}

func (r *objectRepo) All() (*ObjectList, error) {
	where := bson.M{}
	iter := r.gfs.Find(where).Sort("-uploadDate").Iter()
	return r.iterToObjectList(iter), nil
}

func (r *objectRepo) FindByBucket(bucketID bson.ObjectId, skip, limit int) (*ObjectList, error) {
	where := bson.M{"metadata.bucketId": bucketID}
	iter := r.gfs.Find(where).Sort("-uploadDate").Iter()
	return r.iterToObjectList(iter), nil
}

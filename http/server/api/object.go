package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/syb-devs/goose"
	ghttp "github.com/syb-devs/goose/http"
	"gopkg.in/mgo.v2/bson"
)

type reqObjectMetadata struct {
	Title       *string
	Description *string
	Tags        []string
	Custom      map[string]interface{}
}

func (m *reqObjectMetadata) Apply(meta *goose.ObjectMetadata) {
	if m.Title != nil {
		meta.Title = *m.Title
	}
	if m.Description != nil {
		meta.Description = *m.Description
	}
	if m.Tags != nil {
		meta.Tags = m.Tags
	}
	for k, v := range m.Custom {
		meta.Custom[k] = v
	}
}

func listObjects(w http.ResponseWriter, r *http.Request, ctx *ghttp.Context) error {
	bucketID := ctx.URLParams.ByName("bucket")

	bucket, err := getBucketAndCheckAccess(ctx, bucketID, "id", "read")
	if err != nil {
		return ghttp.ProcessError(err)
	}
	repo := goose.NewObjectRepo(ctx.DB)
	olist, err := repo.FindByBucket(bucket.ID, 0, 100)
	if err != nil {
		return ghttp.ProcessError(err)
	}
	defer olist.Close()
	return ghttp.WriteJSON(w, 200, olist.Objects())
}

func listObjectsByIds(w http.ResponseWriter, r *http.Request, ctx *ghttp.Context) error {
	bucketID := ctx.URLParams.ByName("bucket")

	_, err := getBucketAndCheckAccess(ctx, bucketID, "id", "read")
	if err != nil {
		return ghttp.ProcessError(err)
	}
	repo := goose.NewObjectRepo(ctx.DB)

	ids := strings.Split(ctx.URLParams.ByName("objects"), ",")

	oList, err := repo.FindByIds(ids)
	if err != nil {
		return ghttp.ProcessError(err)
	}
	defer oList.Close()
	return ghttp.WriteJSON(w, 200, oList.Objects())
}

func postObject(w http.ResponseWriter, r *http.Request, ctx *ghttp.Context) error {
	fname := ghttp.PrefixSlash(r.URL.Query().Get("name"))
	bucketID := ctx.URLParams.ByName("bucket")
	goose.Log.Debug(fmt.Sprintf("posting object to bucket %s with path %s", bucketID, fname))

	bucket, err := getBucketAndCheckAccess(ctx, bucketID, "id", "write")
	if err != nil {
		return ghttp.ProcessError(err)
	}

	meta := &goose.ObjectMetadata{BucketID: bucket.ID}
	if r.URL.Query().Get("uploaderID") != "" {
		meta.UploaderID = bson.ObjectIdHex(r.URL.Query().Get("uploaderID"))
	}
	repo := goose.NewObjectRepo(ctx.DB)

	var object *goose.Object

	f, fi, err := r.FormFile("object")
	if err == nil {
		// Use the file in the "object" form field
		defer f.Close()
		object, err = repo.Create(f, fname, fi.Header.Get("Content-Type"), meta)
	} else {
		// Use the request body as file data (the RESTful way)
		object, err = repo.Create(r.Body, fname, r.Header.Get("Content-Type"), meta)
	}
	defer object.GridFile().Close()

	resObj, err := repo.OpenId(object.GetID().Hex())
	if err != nil {
		return err
	}
	defer resObj.GridFile().Close()

	return ghttp.WriteJSON(w, 201, resObj)
}

func getObject(w http.ResponseWriter, r *http.Request, ctx *ghttp.Context) error {
	bucketID := ctx.URLParams.ByName("bucket")
	bucket, err := getBucketAndCheckAccess(ctx, bucketID, "id", "read")
	if err != nil {
		return ghttp.ProcessError(err)
	}

	repo := goose.NewObjectRepo(ctx.DB)
	object, err := repo.OpenId(ctx.URLParams.ByName("object"))
	if err != nil {
		return ghttp.ProcessError(err)
	}
	if object.Metadata.BucketID != bucket.ID {
		return ghttp.NewError(404, "")
	}
	return ghttp.WriteJSON(w, 200, object)
}

func deleteObject(w http.ResponseWriter, r *http.Request, ctx *ghttp.Context) error {
	repo := goose.NewObjectRepo(ctx.DB)
	return repo.DeleteId(ctx.URLParams.ByName("object"))
}

func objectFromRequest(r http.Request) (*goose.Object, error) {
	object := &goose.Object{}
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(object)

	return object, err
}

func putObjectMetadata(w http.ResponseWriter, r *http.Request, ctx *ghttp.Context) error {
	repo := goose.NewObjectRepo(ctx.DB)
	object, err := repo.OpenId(ctx.URLParams.ByName("object"))
	if err != nil {
		return ghttp.ProcessError(err)
	}
	reqMeta := &reqObjectMetadata{}
	dec := json.NewDecoder(r.Body)
	err = dec.Decode(reqMeta)
	if err != nil {
		return ghttp.ProcessError(err)
	}
	meta := object.Metadata
	reqMeta.Apply(meta)
	return repo.UpdateMetada(object.ID.Hex(), object.Name, *meta)
}

func getBucketAndCheckAccess(ctx *ghttp.Context, key, keyType, op string) (*goose.Bucket, error) {
	br := goose.NewBucketRepo(ctx.DB)
	var err error
	var bucket *goose.Bucket

	switch keyType {
	case "name":
		bucket, err = br.FindName(key)
	case "id":
		bucket, err = br.FindId(key)
	default:
		return nil, errors.New("invalid value for keyType (accepted are 'name' and 'id'")
	}

	if err != nil {
		return bucket, err
	}

	switch op {
	case "read":
		if !ctx.User.CanWriteBucket(bucket) {
			return bucket, ghttp.ErrForbidden
		}
	case "write":
		if !ctx.User.CanWriteBucket(bucket) {
			return bucket, ghttp.ErrForbidden
		}
	default:
		return bucket, errors.New("invalid value for op (accepted are 'read' and 'write'")
	}
	return bucket, nil
}

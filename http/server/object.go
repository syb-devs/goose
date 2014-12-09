package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"bitbucket.org/syb-devs/goose"
	ghttp "bitbucket.org/syb-devs/goose/http"
)

// ErrNoBucketURL is a 404 Error returned when there's no valid bucket name found in
// the host name (subdomain) or requested URI (first URL part)
var ErrNoBucketURL = ghttp.NewError(404, "invalid bucket in URL")

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

func serveObject(w http.ResponseWriter, r *http.Request, ctx *ghttp.Context) error {
	bucketName, fileName, err := getBucketObjectNames(r)
	goose.Log.Debug(fmt.Sprintf("serving object [%s] from bucket [%s]", fileName, bucketName))
	if err != nil {
		return ghttp.ProcessError(err)
	}

	bucket, err := getBucketAndCheckAccess(ctx, bucketName, "name", "read")
	if err != nil {
		return ghttp.ProcessError(err)
	}

	repo := goose.NewObjectRepo(ctx.DB)
	obj, err := repo.OpenFromBucket(fileName, bucket.ID)
	if err != nil {
		return ghttp.ProcessError(err)
	}

	defer obj.GridFile().Close()
	_, err = io.Copy(w, obj.GridFile())
	return err
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

func postObject(w http.ResponseWriter, r *http.Request, ctx *ghttp.Context) error {
	fname := ghttp.PrefixSlash(r.URL.Query().Get("name"))
	bucketID := ctx.URLParams.ByName("bucket")
	goose.Log.Debug(fmt.Sprintf("posting object to bucket %s with path %s", bucketID, fname))

	bucket, err := getBucketAndCheckAccess(ctx, bucketID, "id", "write")
	if err != nil {
		return ghttp.ProcessError(err)
	}

	meta := &goose.ObjectMetadata{BucketID: bucket.ID}
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

func getBucketObjectNames(r *http.Request) (bucket, object string, err error) {
	subdomain := ghttp.GetSubdomain(r)
	if subdomain == "storage" {
		urlParts := strings.SplitN(r.URL.RequestURI()[1:], "/", 2)
		if len(urlParts) < 2 {
			return "", "", ErrNoBucketURL
		}
		return urlParts[0], ghttp.PrefixSlash(urlParts[1]), nil
	}
	return subdomain, r.URL.RequestURI(), nil
}

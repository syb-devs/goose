package http

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"bitbucket.org/syb-devs/goose"
)

var ErrNoBucketURL = NewError(404, "invalid bucket in URL")

func serveObject(w http.ResponseWriter, r *http.Request, ctx *Context) error {
	bucketName, fileName, err := getBucketObjectNames(r)
	if err != nil {
		return processError(err)
	}

	bucket, err := getBucketAndCheckAccess(ctx, bucketName, "name", "read")
	if err != nil {
		return processError(err)
	}

	repo := goose.NewObjectRepo(ctx.DB)
	obj, err := repo.Open(fileName, bucket.ID)
	if err != nil {
		return processError(err)
	}

	defer obj.Close()
	_, err = io.Copy(w, obj)
	return err
}

func postObject(w http.ResponseWriter, r *http.Request, ctx *Context) error {
	fname := r.URL.Query().Get("name")
	log.Printf("Posting object with path %s", fname)

	bucket, err := getBucketAndCheckAccess(ctx, ctx.URLParams.ByName("bucket"), "id", "write")
	if err != nil {
		return processError(err)
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
	defer object.Close()
	return writeJSON(w, object)
}

func getObject(w http.ResponseWriter, r *http.Request, ctx *Context) error {
	repo := goose.NewObjectRepo(ctx.DB)
	object, err := repo.OpenId(ctx.URLParams.ByName("object"))
	if err != nil {
		return processError(err)
	}
	return writeJSON(w, object)
}

func deleteObject(w http.ResponseWriter, r *http.Request, ctx *Context) error {
	repo := goose.NewObjectRepo(ctx.DB)
	return repo.DeleteId(ctx.URLParams.ByName("object"))
}

func objectFromRequest(r http.Request) (*goose.Object, error) {
	object := &goose.Object{}
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(object)

	return object, err
}

func postObjectMetadata(w http.ResponseWriter, r *http.Request, ctx *Context) error {
	repo := goose.NewObjectRepo(ctx.DB)
	object, err := repo.OpenId(ctx.URLParams.ByName("object"))
	if err != nil {
		return processError(err)
	}
	meta := object.Metadata()
	dec := json.NewDecoder(r.Body)
	err = dec.Decode(meta)
	if err != nil {
		return processError(err)
	}
	return repo.UpdateMetada(object.Id().Hex(), object.Name(), *meta)
}

func getBucketAndCheckAccess(ctx *Context, key, keyType, op string) (*goose.Bucket, error) {
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
			return bucket, ErrForbidden
		}
	case "write":
		if !ctx.User.CanWriteBucket(bucket) {
			return bucket, ErrForbidden
		}
	default:
		return bucket, errors.New("invalid value for op (accepted are 'read' and 'write'")
	}
	return bucket, nil
}

func getBucketObjectNames(r *http.Request) (bucket, object string, err error) {
	subdomain := getSubdomain(r)
	if subdomain == "storage" {
		urlParts := strings.SplitN(r.URL.RequestURI()[1:], "/", 2)
		if len(urlParts) < 2 {
			return "", "", ErrNoBucketURL
		}
		return urlParts[0], urlParts[1], nil
	}
	return subdomain, r.URL.RequestURI(), nil
}

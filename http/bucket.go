package http

import (
	"encoding/json"
	"net/http"

	"bitbucket.org/syb-devs/goose"
)

var ErrBucketExists = NewError(409, "the bucket already exists")

func postBucket(w http.ResponseWriter, r *http.Request, ctx *Context) error {
	//TODO: validation of the POSTed data
	repo := goose.NewBucketRepo(ctx.DB)
	bucket, err := bucketFromRequest(*r)
	if err != nil {
		return err
	}
	if repo.Exists(bucket.Name) {
		return ErrBucketExists
	}
	err = repo.Insert(bucket)
	if err != nil {
		return err
	}
	return writeJSON(w, bucket)
}

func getBucket(w http.ResponseWriter, r *http.Request, ctx *Context) error {
	repo := goose.NewBucketRepo(ctx.DB)
	bucket, err := repo.FindId(ctx.URLParams.ByName("bucket"))
	if err != nil {
		return processError(err)
	}
	if !ctx.User.CanReadBucket(bucket) {
		return ErrForbidden
	}
	return writeJSON(w, bucket)
}

func putBucket(w http.ResponseWriter, r *http.Request, ctx *Context) error {
	repo := goose.NewBucketRepo(ctx.DB)
	bucket, err := repo.FindId(ctx.URLParams.ByName("bucket"))
	if err != nil {
		return processError(err)
	}
	if !ctx.User.CanWriteBucket(bucket) {
		return ErrForbidden
	}
	dec := json.NewDecoder(r.Body)
	err = dec.Decode(bucket)
	if err != nil {
		return err
	}
	if err = repo.Update(bucket); err != nil {
		return processError(err)
	}
	return writeJSON(w, bucket)
}

func deleteBucket(w http.ResponseWriter, r *http.Request, ctx *Context) error {
	repo := goose.NewBucketRepo(ctx.DB)
	bucket, err := repo.FindId(ctx.URLParams.ByName("bucket"))
	if err != nil {
		return processError(err)
	}
	if !ctx.User.CanWriteBucket(bucket) {
		return ErrForbidden
	}
	return processError(repo.DeleteId(ctx.URLParams.ByName("bucket")))
}

func bucketFromRequest(r http.Request) (*goose.Bucket, error) {
	bucket := &goose.Bucket{}
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(bucket)

	return bucket, err
}

package main

import (
	"encoding/json"
	"net/http"

	"bitbucket.org/syb-devs/goose"
	ghttp "bitbucket.org/syb-devs/goose/http"
)

// ErrBucketExists represents an HTTP 409 error, returned when the user is trying to create a bucket,
// but another with the same name already exists
var ErrBucketExists = ghttp.NewError(409, "the bucket already exists")

type reqBucket struct {
	Name *string
}

func (b *reqBucket) Apply(bucket *goose.Bucket) {
	if b.Name != nil {
		bucket.Name = *b.Name
	}
}

func postBucket(w http.ResponseWriter, r *http.Request, ctx *ghttp.Context) error {
	//TODO: validation of the POSTed data
	repo := goose.NewBucketRepo(ctx.DB)
	reqBucket, err := bucketFromRequest(*r)
	if err != nil {
		return err
	}
	if repo.Exists(*reqBucket.Name) {
		return ErrBucketExists
	}
	bucket := &goose.Bucket{}
	reqBucket.Apply(bucket)
	err = repo.Insert(bucket)
	if err != nil {
		return err
	}
	return ghttp.WriteJSON(w, bucket)
}

func getBucket(w http.ResponseWriter, r *http.Request, ctx *ghttp.Context) error {
	repo := goose.NewBucketRepo(ctx.DB)
	bucket, err := repo.FindId(ctx.URLParams.ByName("bucket"))
	if err != nil {
		return ghttp.ProcessError(err)
	}
	if !ctx.User.CanReadBucket(bucket) {
		return ghttp.ErrForbidden
	}
	return ghttp.WriteJSON(w, bucket)
}

func putBucket(w http.ResponseWriter, r *http.Request, ctx *ghttp.Context) error {
	repo := goose.NewBucketRepo(ctx.DB)
	bucket, err := repo.FindId(ctx.URLParams.ByName("bucket"))
	if err != nil {
		return ghttp.ProcessError(err)
	}
	if !ctx.User.CanWriteBucket(bucket) {
		return ghttp.ErrForbidden
	}
	dec := json.NewDecoder(r.Body)
	var reqBucket *reqBucket
	err = dec.Decode(reqBucket)
	if err != nil {
		return err
	}
	reqBucket.Apply(bucket)

	if err = repo.Update(bucket); err != nil {
		return ghttp.ProcessError(err)
	}
	return ghttp.WriteJSON(w, bucket)
}

func deleteBucket(w http.ResponseWriter, r *http.Request, ctx *ghttp.Context) error {
	repo := goose.NewBucketRepo(ctx.DB)
	bucket, err := repo.FindId(ctx.URLParams.ByName("bucket"))
	if err != nil {
		return ghttp.ProcessError(err)
	}
	if !ctx.User.CanWriteBucket(bucket) {
		return ghttp.ErrForbidden
	}
	return ghttp.ProcessError(repo.DeleteId(ctx.URLParams.ByName("bucket")))
}

func bucketFromRequest(r http.Request) (*reqBucket, error) {
	bucket := &reqBucket{}
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(bucket)

	return bucket, err
}

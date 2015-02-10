package main

import (
	"github.com/dimfeld/httptreemux"
	ghttp "github.com/syb-devs/goose/http"
)

func newRouter() *httptreemux.TreeMux {
	rt := httptreemux.New()
	ctx := ghttp.HandlerAdapterTreeMux

	rt.POST("/buckets", ctx(postBucket))
	rt.GET("/buckets/:bucket", ctx(getBucket))
	rt.GET("/buckets/name/:bucket", ctx(getBucketByName))
	rt.PUT("/buckets/:bucket", ctx(putBucket))
	rt.DELETE("/buckets/:bucket", ctx(deleteBucket))

	rt.GET("/buckets/:bucket/objects", ctx(listObjects))
	rt.GET("/buckets/:bucket/objects/list/:objects", ctx(listObjectsByIds))
	rt.POST("/buckets/:bucket/objects", ctx(postObject))
	rt.GET("/buckets/:bucket/objects/:object", ctx(getObject))
	rt.DELETE("/buckets/:bucket/objects/:object", ctx(deleteObject))

	rt.PUT("/buckets/:bucket/objects/:object/metadata", ctx(putObjectMetadata))

	return rt
}

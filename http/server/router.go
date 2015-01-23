package main

import (
	"net/http"
	"strings"

	"github.com/dimfeld/httptreemux"
	ghttp "github.com/syb-devs/goose/http"
)

type router struct {
	*httptreemux.TreeMux
}

func newRouter() *router {
	return &router{TreeMux: httptreemux.New()}
}

func (rt *router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if rt.isAPI(r) {
		rt.TreeMux.ServeHTTP(w, r)
		return
	}
	ghttp.HandlerAdapter(serveObject)(w, r, nil)
}

func (rt *router) isAPI(r *http.Request) bool {
	//TODO: enable customisation
	return strings.HasPrefix(r.Host, "api.")
}

func (rt *router) withRoutes() *router {
	ctx := ghttp.HandlerAdapter

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

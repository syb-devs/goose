package http

import (
	"net/http"
	"strings"

	"github.com/dimfeld/httptreemux"
)

func NewRouter() *Router {
	return &Router{TreeMux: httptreemux.New()}
}

func (rt *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if rt.isAPI(r) {
		rt.TreeMux.ServeHTTP(w, r)
		return
	}
	HandlerAdapter(serveObject)(w, r, nil)
}

func (rt *Router) isAPI(r *http.Request) bool {
	//TODO: enable customisation
	return strings.HasPrefix(r.Host, "api.")
}

func (rt *Router) WithRoutes() *Router {
	ctx := HandlerAdapter

	rt.POST("/buckets", ctx(postBucket))
	rt.GET("/buckets/:bucket", ctx(getBucket))
	rt.PUT("/buckets/:bucket", ctx(putBucket))
	rt.DELETE("/buckets/:bucket", ctx(deleteBucket))

	rt.POST("/buckets/:bucket/objects", ctx(postObject))
	rt.GET("/buckets/:bucket/objects/:object", ctx(getObject))
	rt.DELETE("/buckets/:bucket/objects/:object", ctx(deleteObject))

	rt.PUT("/buckets/:bucket/objects/:object/metadata", ctx(postObjectMetadata))

	return rt
}

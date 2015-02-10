package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/syb-devs/dockerlink"
	"github.com/syb-devs/goose"
	ghttp "github.com/syb-devs/goose/http"
)

// ErrNoBucketURL is a 404 Error returned when there's no valid bucket name found in
// the host name (subdomain) or requested URI (first URL part)
var ErrNoBucketURL = ghttp.NewError(404, "invalid bucket in URL")

func main() {
	goose.NewDBConn(goose.DBOptions{
		Database:     envDefault("DBNAME", "goose"),
		URL:          getMongoURI(),
		SetAsDefault: true,
	})

	addr := fmt.Sprintf(":%s", envDefault("PORT", "80"))
	ctx := ghttp.HandlerAdapter
	log.Fatal(http.ListenAndServe(addr, http.HandlerFunc(ctx(serveObject))))
}

func envDefault(key, defval string) string {
	val := os.Getenv(key)
	if val != "" {
		return val
	}
	return defval
}

func getMongoURI() string {
	if uri := os.Getenv("MONGO_URL"); uri != "" {
		return uri
	}
	if link, err := dockerlink.GetLink(envDefault("DOCKER_MONGO_NAME", "mongodb"), 27017, "tcp"); err == nil {
		return fmt.Sprintf("%s:%d", link.Address, link.Port)
	}
	panic("mongodb connection not found, use MONGO_URL env var or a docker link with mongodb name")
}

func serveObject(w http.ResponseWriter, r *http.Request, ctx *ghttp.Context) error {
	bucketName, fileName, err := getBucketObjectNames(r)
	goose.Log.Debug(fmt.Sprintf("serving object [%s] from bucket [%s]", fileName, bucketName))
	if err != nil {
		return ghttp.ProcessError(err)
	}

	br := goose.NewBucketRepo(ctx.DB)
	bucket, err := br.FindName(bucketName)

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

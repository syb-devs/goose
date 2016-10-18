package main

import (
	"fmt"
	"net/http"
	"os"
	"runtime/debug"

	ghttp "github.com/syb-devs/goose/http"
	"github.com/syb-devs/goose/http/client"
)

func main() {
	defer panicHandler()

	bucketName := "mybucket"

	// Setup the service
	print("setting up the API client...")
	service, err := client.New(&http.Client{}, "http://localhost:3000")
	handle(err)

	// Get the bucket
	print("retrieving the bucket %v...", bucketName)
	bucket, err := service.Buckets.RetrieveByName(bucketName)
	httpErr, ok := err.(*ghttp.Error)
	if ok && httpErr.Code == 404 {
		print("the bucket does not exist, let's create it!")
		// Create the bucket if it does not exist
		bucket, err = service.Buckets.Create(bucketName)
		handle(err)
		print("\nbucket created: %+v", bucket)
	} else {
		handle(err)
		print("\nbucket retrieved: %+v", bucket)
	}

	// Upload a file to the bucket
	print("\nuploading file...")
	file, err := os.Open("tesla_colorado.jpg")
	handle(err)

	object, err := service.Objects.Upload(bucket.ID.Hex(), "/uploads/tesla colorado.jpg", "image/jpeg", file, nil)
	handle(err)
	print("\nfile uploaded: %+v", object)

	// List all objects in the bucket
	print("\nlisting bucket files...")
	objects, err := service.Objects.List(bucket.ID.Hex())
	handle(err)
	print("\nbucket objects: %+v", objects)

	print("\ndeleting uploaded file...")
	// Delete the uploaded file
	err = service.Objects.Delete(bucket.ID.Hex(), object.ID.Hex())
	handle(err)
}

func handle(err error) {
	if err != nil {
		print("error: %T", err)
		panic(err)
	}
}

func print(pattern string, tokens ...interface{}) (n int, err error) {
	return fmt.Printf("\n"+pattern, tokens...)
}

func panicHandler() {
	if p := recover(); p != nil {
		print("recovering from panic: %v\nstack trace: %s", p, debug.Stack())
	}
}

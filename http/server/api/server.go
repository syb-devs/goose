package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/syb-devs/dockerlink"
	"github.com/syb-devs/goose"
)

func main() {
	goose.NewDBConn(goose.DBOptions{
		Database:     envDefault("DBNAME", "goose"),
		URL:          getMongoURI(),
		SetAsDefault: true,
	})

	addr := fmt.Sprintf(":%s", envDefault("PORT", "8080"))

	log.Printf("Goose API server listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, newRouter()))
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

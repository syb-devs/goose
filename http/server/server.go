package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/syb-devs/goose"
)

func main() {
	goose.NewDBConn(goose.DBOptions{
		Database:     envDefault("DBNAME", "goose"),
		URL:          envDefault("MONGOURL", "localhost"),
		SetAsDefault: true,
	})

	addr := fmt.Sprintf(":%s", envDefault("PORT", "8080"))
	log.Fatal(http.ListenAndServe(addr, newRouter().withRoutes()))
}

func envDefault(key, defval string) string {
	val := os.Getenv(key)
	if val != "" {
		return val
	}
	return defval
}

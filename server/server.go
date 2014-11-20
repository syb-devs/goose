package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"bitbucket.org/syb-devs/goose"
	ghttp "bitbucket.org/syb-devs/goose/http"
)

func main() {
	addr := fmt.Sprintf(":%s", os.Getenv("PORT"))

	goose.NewDBConn(goose.DBOptions{Database: "goose", URL: "172.17.0.2", SetAsDefault: true})

	log.Fatal(http.ListenAndServe(addr, ghttp.NewRouter().WithRoutes()))
}

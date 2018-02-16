package main

import (
	"flag"
	"net/http"

	_ "github.com/gyf1214/chaty/controller"
	"github.com/gyf1214/chaty/model"
)

var (
	listen = flag.String("listen", ":12450", "listen port")
)

func main() {
	flag.Parse()
	err := model.LoadChannels()
	if err != nil {
		panic(err)
	}

	http.ListenAndServe(*listen, nil)
}

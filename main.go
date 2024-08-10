package main

import (
	"fmt"
	"github.com/xionghengheng/ff_plib/db"
	"log"
	"net/http"
)

func main() {
	if err := db.Init(); err != nil {
		panic(fmt.Sprintf("mysql init failed with %+v", err))
	}

	//统一上报
	http.HandleFunc("/api/report", Report)

	log.Fatal(http.ListenAndServe(":80", nil))
}

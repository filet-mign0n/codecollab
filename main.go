package main

import (
	"log"
	"flag"
	"net/http"
)

var host = flag.String("h", "localhost", "host")
var port = flag.String("p", "8000", "port")

// logging level flags
var debug = flag.Bool("d", false, "debug")
var verbose = flag.Bool("v", false, "verbose")

func logDebug(s string) {
	if *debug {
		log.Println(s)
	}
}
func logVerbose(s string) {
	if *verbose {
		log.Println(s)
	}
}

func main() {

	flag.Parse()

	h := NewHub()
	go h.run()

	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ServeWs(w, r, h)
	})

	log.Println("CodeCollab server listening on", *host+":"+*port)
	log.Println(http.ListenAndServe(*host+":"+*port, nil))

}

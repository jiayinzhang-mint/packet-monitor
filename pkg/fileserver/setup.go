package fileserver

import (
	"io"
	"log"
	"net/http"
	"net/http/pprof"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func Init() {
	var (
		port = "8100"
	)

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", startPage).Methods("GET")
	router.PathPrefix("/export/").Handler(http.StripPrefix("/export/", http.FileServer(http.Dir("./export"))))

	router.HandleFunc("/debug/pprof/", pprof.Index)
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	router.HandleFunc("/debug/pprof/trace", pprof.Trace)

	router.Handle("/debug/pprof/allocs", pprof.Handler("allocs"))
	router.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	router.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	router.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	router.Handle("/debug/pprof/block", pprof.Handler("block"))

	logrus.Printf("Serving on HTTP port: %s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))

}

func startPage(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "dimoni start page")
}

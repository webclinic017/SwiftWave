package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/felixge/fgprof"
)

func main() {
	profilingEnabled := os.Getenv("PROFILING")
	if profilingEnabled == "1" {
		go func() {
			http.DefaultServeMux.Handle("/debug/fgprof", fgprof.Handler())
			log.Println(http.ListenAndServe(":6060", nil))
		}()
	}

	if err := CreateDatabaseFileIfNotExist(); err != nil {
		panic(err)
	}
	if err := InitiateDatabaseInstances(); err != nil {
		panic(err)
	}
	if err := SetupIptablesChains(); err != nil {
		panic(err)
	}
	rootCmd.Execute()
}

// Copyright 2022 by lolorenzo77. All rights reserved.
// Use of this source code is governed by MIT licence that can be found in the LICENSE file.

// Single Page Application (SPA) Web Server in go
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/sunraylab/spaserver/configs"
	"github.com/sunraylab/verbose"
)

func main() {
	// handle verbose flag
	fverbose := flag.Bool("verbose", false, "verbose output")
	flag.Parse()
	verbose.IsOn = *fverbose

	cfg, err := configs.LoadConfiguration(os.Getenv("SPA_ENV"))
	if err != nil {
		fmt.Println("unable to start the server")
		os.Exit(1)
	}

	// let's go
	fmt.Printf("Starting the SPA web server serving pages and APIs on port %s\n", cfg.HttpPort)
	counter := 0

	// configure the server
	webrouter := mux.NewRouter()
	// with or without trailing slash is the same route
	webrouter.StrictSlash(true)
	// an example API handler. add your api routes here after.
	webrouter.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		counter++
		json.NewEncoder(w).Encode(map[string]string{"health": "live", "counter": strconv.Itoa(counter)})
	})
	// the main handler
	webrouter.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// force header for wasm files
		if strings.HasSuffix(r.URL.Path, ".wasm") {
			w.Header().Set("content-type", "application/wasm")
		}
		// serve spa files
		http.FileServer(http.Dir(cfg.SpaDir)).ServeHTTP(w, r)
	})

	// add middleware to log every request if in verbose mode
	if verbose.IsOn {
		fmt.Println("logging is on")
		webrouter.Use(middlewareLogging)
	}
	// add middleware to remove cache if requested in config file
	if !cfg.HttpCacheControl {
		fmt.Println("cache is off")
		webrouter.Use(middlewareNoCache)
	}

	// setup timeout
	srv := &http.Server{
		Handler:      webrouter,
		Addr:         cfg.HttpPort,
		WriteTimeout: cfg.HttpRWTimeout * time.Second,
		ReadTimeout:  cfg.HttpRWTimeout * time.Second,
		IdleTimeout:  cfg.HttpIdleTimeout * time.Second,
	}

	// listen and serve in a go routine to allow catching shutdown request in parallel
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL will not be caught.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	// Block until we receive a shutdown signal.
	<-c

	// Start the clean shutdown process.
	// Create a deadline to wait for, longer than the rwTimeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.HttpRWTimeout+time.Second*10)
	defer cancel()

	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)

	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	fmt.Println("SPA web server shutdown")
	os.Exit(0)
}

/*
   middlewares
*/

func middlewareNoCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Remove cache in the answer
		w.Header().Add("Cache-Control", "no-cache")
		// Call the next handler (another middleware in the chain or the final handler)
		next.ServeHTTP(w, r)
	})
}

func middlewareLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// log every request
		log.Println(r.RequestURI)
		// Call the next handler (another middleware in the chain or the final handler)
		next.ServeHTTP(w, r)
	})
}

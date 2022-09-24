// Single Page Application (SPA) Web Server in go
package main

import (
	"log"
	"net/http"
	"strings"
)

func main() {

	const staticdir = "./webapp"
	const port = ":5500"

	fs := http.FileServer(http.Dir(staticdir))

	log.Printf("Serving %s on http://localhost%s\n", staticdir, port)

	log.Fatal(http.ListenAndServe(port, http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		// for dev, unset cache
		resp.Header().Add("Cache-Control", "no-cache")

		// apply a specific header for .wasm file
		if strings.HasSuffix(req.URL.Path, ".wasm") {
			resp.Header().Set("content-type", "application/wasm")
		}

		fs.ServeHTTP(resp, req)
	})))
}

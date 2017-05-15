package main

import (
	"fmt"
	"net/http"
	"time"
	"os"
	"strings"
)

var contentDir string = "./app"

func init() {
  	if len(os.Args) > 1 {
  		contentDir = os.Args[1]
  	}

}

func main() {
	addr := "0.0.0.0:8080"
	serveSite(addr)
}

func serveSite(listenAddr string) {
	// serve the 'site'
	http.Handle("/hello", logHandler(helloHandler()))              // says hello
	http.Handle("/", logHandler(http.FileServer(http.Dir(contentDir)))) // serve static files

	// start server
	fmt.Printf("Starting server on %s...\n", listenAddr)
	err := http.ListenAndServe(listenAddr, nil)
	if err != nil {
		fmt.Printf("Failed to start server - %s\n", err)
	}
}

// hello's handler (GET only)
func helloHandler() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			rw.WriteHeader(http.StatusNotFound)
			return
		}

		hello(rw, req)
	})
}

// hello says hi
func hello(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("hello there!\n"))
}

// logging middleware
func logHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		start := time.Now()

		writer := &LogRW{rw, 0, 0}
		h.ServeHTTP(writer, req)
		// if writer.status != http.StatusOK {
		if writer.status == http.StatusNotFound {
			writer.CustomErr(rw, req, writer.status)
		}

		remoteAddr := req.RemoteAddr
		if fwdFor := req.Header.Get("X-Forwarded-For"); len(fwdFor) > 0 {
			// get actual remote (last is oldest remote addr)
			fwds := strings.Split(string(fwdFor), ",")
			remoteAddr = strings.Trim(fwds[len(fwds)-1], " ")
		} 

		fmt.Printf("%s %s %s%s %s %d(%d) - %s [User-Agent: %s] (%s)\n",
			time.Now().Format(time.RFC3339), req.Method, req.Host, req.RequestURI,
			req.Proto, writer.status, writer.wrote, remoteAddr,
			req.Header.Get("User-Agent"), time.Since(start))
	})
}

// LogRW is provides the logging functionality i've always wanted, giving access
// to the number bytes written, as well as the status. (I try to always writeheader
// prior to write, so status works fine for me)
type LogRW struct {
	http.ResponseWriter
	status int
	wrote  int
}

// WriteHeader matches the response writer interface, and stores the status
func (n *LogRW) WriteHeader(status int) {
	n.status = status

	// http.FileServer and its (http.)Error() function will write text/plain headers
	// which cause the browser to not render the html from our custom error page.
	// write 404 page to current url rather than redirect so refreshing the page will
	// work properly (if the page becomes available later)
	if status != 404 {
		n.ResponseWriter.WriteHeader(status)
	}
}

// Write matches the response writer interface, and stores the number of bytes written
func (n *LogRW) Write(p []byte) (int, error) {
	if n.status == http.StatusNotFound {
		n.wrote = len(p)
		return n.wrote, nil
	}
	wrote, err := n.ResponseWriter.Write(p)
	n.wrote = wrote
	return wrote, err
}

// CustomErr allows us to write a custom error file to the user. It is part of
// LogRW so we can track the bytes written.
func (n *LogRW) CustomErr(w http.ResponseWriter, r *http.Request, status int) {
	if status == http.StatusNotFound {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(http.StatusNotFound)

		n.wrote, _ = w.Write([]byte("Not found\n"))
	}
}

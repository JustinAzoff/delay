package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

func DelayServer(w http.ResponseWriter, r *http.Request) {
	var keepalive time.Duration
	count := uint64(1)
	keepalive = -1

	// Delay is required
	delayString := r.FormValue("delay")
	delay, err := strconv.ParseUint(delayString, 10, 0)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid or missing delay argument\n")
		return
	}

	// keepalive is optional, defaults to disabled
	keepAliveString := r.FormValue("keepalive")
	if keepAliveString != "" {
		keepaliveSeconds, err := strconv.ParseUint(keepAliveString, 10, 0)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Invalid keepalive argument\n")
			return
		}
		keepalive = time.Duration(keepaliveSeconds) * time.Second
	}

	// count is optional, defaults to 1
	countString := r.FormValue("count")
	if countString != "" {
		count, err = strconv.ParseUint(countString, 10, 0)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Invalid count argument\n")
			return
		}
	}

	log.Printf("GET: %s delay: %v. keepalive: %v. count: %v", r.RemoteAddr, delay, keepalive, count)

	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	conn, rw, err := hj.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		log.Println("Not a TCP connection")
		return
	}
	if keepalive != -1 {
		err = tcpConn.SetKeepAlivePeriod(keepalive)
		if err != nil {
			log.Println("Failed to SetKeepAlivePeriod: %v", err)
		}
	} else {
		err = tcpConn.SetKeepAlive(false)
		if err != nil {
			log.Println("Failed to SetKeepAlive: %v", err)
		}
	}

	// At this point we hijacked the socket so we need to speak http ourself.

	fmt.Fprintf(rw, "HTTP/1.1 200 OK\r\nContent-Type: text/event-stream\r\nConnection: close\r\n\r\n")
	rw.Flush()

	for i := range count {
		time.Sleep(time.Duration(delay) * time.Second)
		fmt.Fprintf(rw, "# %d %s Delayed for %d seconds\r\n", i+1, time.Now().Format(time.DateTime), delay)
		err := rw.Flush()
		if err != nil {
			log.Printf("Failed to Flush: %v", err)
			return
		}
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", DelayServer)
	log.Println("listening on", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

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
	delayString := r.PathValue("delay")
	delay, err := strconv.Atoi(delayString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid or missing delay argument\n")
		return
	}

	var keepalive time.Duration
	keepalive = -1

	keepAliveString := r.FormValue("keepalive")
	if keepAliveString != "" {
		keepaliveSeconds, err := strconv.ParseUint(keepAliveString, 10, 0)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Invalid keepalive argument\n")
		}
		keepalive = time.Duration(keepaliveSeconds) * time.Second
	}
	log.Printf("GET: %s delay: %v. keepalive: %v", r.RemoteAddr, delay, keepalive)

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

	// You can now access the underlying net.Conn and its methods
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

	time.Sleep(time.Duration(delay) * time.Second)
	fmt.Fprintf(rw, "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\nDelayed for %d seconds\r\n", delay)
	rw.Flush()
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/delay/{delay}", DelayServer)
	log.Println("listening on", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

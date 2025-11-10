package main

import (
    "fmt"
    "log"
    "net/http"
    "os"
)

func main() {
    port := os.Getenv("SERVICE_PORT")
    if port == "" {
        port = "7302"
    }
    mux := http.NewServeMux()
    mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(200)
        w.Write([]byte("ok"))
    })
    srv := &http.Server{Addr: ":"+port, Handler: mux}
    log.Printf("perf-coach-service listening on :%s", port)
    if err := srv.ListenAndServe(); err != nil {
        fmt.Println(err)
    }
}

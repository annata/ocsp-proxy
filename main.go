package main

import (
        "flag"
        "os"
        "net/http"
        "log"
        "net/http/httputil"
)

func main() {
        var addr string
        var ocspHost string

        flag.StringVar(&addr, "http", "", "HTTP host:port to listen to")
        flag.StringVar(&ocspHost, "ocsphost", "", "OCSP server to proxy requests to")
        flag.Parse()

        if ocspHost == "" {
            ocspHost = os.Getenv("ocsphost")
        }
        if ocspHost == "" {
                log.Fatal("need ocsphost parameter")
        }

        if addr == "" {
            addr = os.Getenv("http")
        }
        if addr == "" {
            addr = ":8080"
        }

        rp := &httputil.ReverseProxy{
                Director: func(req *http.Request) {
                        req.URL.Scheme = "http"
                        req.URL.Host = ocspHost
                        req.Host = req.URL.Host
                        //log.Printf("forward to %s", req.URL.Host)
                },
                Transport: http.DefaultTransport,
        }
        log.Printf("Serving ocsp proxy on %s\n", addr)
        log.Printf("Forward ocsp requests to the host: %s\n", ocspHost)
        http.ListenAndServe(addr, rp)
}

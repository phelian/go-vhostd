package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/phelian/log"
)

type vhost struct {
	Vhost string `json:"vhost"`
	Host  string `json:"host"`
}

type config struct {
	Host    string        `json:"host"`
	Port    int           `json:"port"`
	Timeout time.Duration `json:"timeout"`
	Log     log.Config    `json:"log"`
	Vhosts  []vhost       `json:"vhosts"`
}

func main() {
	if len(os.Args) != 2 {
		usage()
		os.Exit(1)
	}

	cfg, err := readConfig(os.Args[1])
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}

	logger, err := log.New(cfg.Log)
	if err != nil {
		log.Println(err.Error())
	}

	run(cfg, logger)
}

func usage() {
	fmt.Println("Usage go-vhostd <config>")
}

func readConfig(path string) (*config, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cfg := &config{}
	err = json.Unmarshal(file, cfg)
	return cfg, err
}

func proxyHandler(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Header.Add("Proxy-RemoteIP", strings.Split(r.RemoteAddr, ":")[0])
		p.ServeHTTP(w, r)
	}
}

func run(cfg *config, logger *log.Handle) {
	mux := http.NewServeMux()

	for _, vhost := range cfg.Vhosts {
		urlString, err := url.Parse("http://" + vhost.Host)
		if err != nil {
			panic(err)
		}
		proxy := httputil.NewSingleHostReverseProxy(urlString)
		log.Printf("Setting up redirection for %s to %s\n", vhost.Vhost, vhost.Host)
		mux.HandleFunc(vhost.Vhost+"/", proxyHandler(proxy))
	}

	// Server setup
	address := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	server := &http.Server{
		Addr:           address,
		Handler:        logHTTP(mux, logger),
		ReadTimeout:    cfg.Timeout * time.Second,
		WriteTimeout:   cfg.Timeout * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Start server
	log.Printf("Listening to http://%s\n", address)

	panic(server.ListenAndServe())
}

func logHTTP(handler http.Handler, logger *log.Handle) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Infof("Request: %s %s %s", r.RemoteAddr, r.Method, r.Host+r.URL.String())
		handler.ServeHTTP(w, r)
	})
}

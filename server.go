package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"gopkg.in/redis.v3"
)

var (
	fProject = flag.String("project", "site", "project to fetch from redis for")

	fAddress  = flag.String("address", ":4000", "address to listen on")
	fRedis    = flag.String("redis", "localhost:6379", "address to connect to redis")
	fPassword = flag.String("password", "", "redis password to use")
	fDB       = flag.Int("db", 0, "redis db to use")

	fTLSAddres = flag.String("tls-address", "", "tls address to listen on")
	fTLSKey    = flag.String("tls-key", "", "tls key")
	fTLSCert   = flag.String("tls-cert", "", "tls cert")
)

type LightningHandler struct {
	r              *redis.Client
	project        string
	currentContent string
}

func (l *LightningHandler) Connect(addr, pass string, db int) error {
	l.r = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pass,
		DB:       int64(db),
	})

	_, err := l.r.Ping().Result()
	return err
}

func (l *LightningHandler) SetProject(proj string) {
	l.project = proj
	l.currentContent = fmt.Sprintf("%s:index:current-content", proj)
}

func (l *LightningHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	currentContent := l.currentContent

	index := req.URL.Query().Get("index_key")
	if index != "" {
		currentContent = fmt.Sprintf("%s:index:%s", l.project, index)
	}

	str, err := l.r.Get(currentContent).Result()
	if err != nil {
		res.WriteHeader(500)
		return
	}

	res.Header().Add("Content-Type", "text/html")
	res.Write([]byte(str))
}

func main() {
	flag.Parse()

	var handler LightningHandler

	handler.SetProject(*fProject)

	err := handler.Connect(*fRedis, *fPassword, *fDB)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Lightning server booted.\nListening on %s\n", *fAddress)

	if *fTLSAddres != "" {
		fmt.Printf("Listening on TLS %s\n", *fTLSAddres)

		go func() {
			err := http.ListenAndServeTLS(*fTLSAddres, *fTLSCert, *fTLSKey, &handler)
			if err != nil {
				log.Fatal(err)
			}
		}()
	}

	err = http.ListenAndServe(*fAddress, &handler)
	if err != nil {
		log.Fatal(err)
	}
}

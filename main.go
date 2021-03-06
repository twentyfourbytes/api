package main

import (
	"encoding/json"
	"expvar"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/logger"

	"github.com/go-redis/redis"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/twentyfourbytes/api/config"
)

var redisClient *redis.Client

func main() {
	configuration := config.Config()

	lf, err := os.OpenFile("log.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		logger.Fatalf("Failed to open log file: %v", err)
	}
	defer lf.Close()
	defer logger.Init("API", true, false, lf).Close()

	// initialise redis
	redisClient = redis.NewClient(&redis.Options{
		Addr: configuration.Redis.IP + ":" + strconv.Itoa(configuration.Redis.Port),
		DB:   0,
	})

	corsObj := handlers.AllowedOrigins([]string{"*"})

	r := mux.NewRouter()
	r.HandleFunc("/dns", getDNS)
	r.HandleFunc("/ip", getIP)
	r.HandleFunc("/download", download)
	r.HandleFunc("/upload", upload)
	r.HandleFunc("/speeds", getSpeeds)
	r.HandleFunc("/speed", getSpeed)

	r.HandleFunc("/debug/vars", expvarHandler)
	// r.HandleFunc("/upload", handleUpload)

	http.Handle("/", r)

	srv := &http.Server{
		Handler:      handlers.CORS(corsObj)(r),
		Addr:         configuration.Server,
		WriteTimeout: 15 * time.Minute,
		ReadTimeout:  15 * time.Minute,
	}
	logger.Infoln("Server running at", configuration.Server)
	logger.Errorln("Error testing upload speed ")

	logger.Fatal(srv.ListenAndServe())
}

func getDNS(w http.ResponseWriter, r *http.Request) {
	var dnsServerips []string
	vars := r.URL.Query().Get("r")
	logger.Infoln("Requested dns query for key ", vars)

	dnsip := redisClient.Keys(vars + ":*")
	for _, v := range dnsip.Val() {
		val := redisClient.Get(v)
		err := val.Err()
		if err != nil {
			logger.Errorln("Error getting redis key", err)
		}
		dnsServerips = append(dnsServerips, val.Val())
	}
	w.WriteHeader(http.StatusOK)
	js, e := json.Marshal(dnsServerips)
	if e != nil {
		logger.Errorln(e)
	}
	w.Write(js)
}

func getSpeeds(w http.ResponseWriter, r *http.Request) {
	var speeds []Speed
	vars := r.URL.Query().Get("r")
	logger.Infoln("Requested All speeds for key ", vars)

	dnsip := redisClient.Keys(vars + "*")
	for _, v := range dnsip.Val() {
		val := redisClient.Get(v)
		var s Speed
		b, _ := val.Bytes()
		json.Unmarshal(b, &s)
		speeds = append(speeds, s)
	}
	w.WriteHeader(http.StatusOK)
	js, e := json.Marshal(speeds)
	if e != nil {
		logger.Errorln(e)
	}
	w.Write(js)
}
func getSpeed(w http.ResponseWriter, r *http.Request) {
	vars := r.URL.Query().Get("r")
	logger.Infoln("Requested speeds for key ", vars)
	val := redisClient.Get(vars)
	var s Speed
	b, _ := val.Bytes()
	json.Unmarshal(b, &s)
	w.WriteHeader(http.StatusOK)
	js, e := json.Marshal(s)
	if e != nil {
		logger.Errorln(e)
	}
	w.Write(js)
}

func getIP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	js, e := json.Marshal(strings.Split(r.RemoteAddr, ":")[0])
	if e != nil {
		logger.Errorln(e)
	}
	w.Write(js)
}

func expvarHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(w, "{\n")
	first := true
	expvar.Do(func(kv expvar.KeyValue) {
		if !first {
			fmt.Fprintf(w, ",\n")
		}
		first = false
		fmt.Fprintf(w, "%q: %s", kv.Key, kv.Value)
	})
	fmt.Fprintf(w, "\n}\n")
}

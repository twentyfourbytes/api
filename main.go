package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/logger"

	"github.com/go-redis/redis"

	"github.com/gorilla/mux"
)

var redisClient *redis.Client

type Configuration struct {
	Redis struct {
		Port int    `json:"port"`
		Ip   string `json:"ip"`
	} `json:"redis"`
	Server string `json:"server"`
}

func main() {

	lf, err := os.OpenFile("log.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		logger.Fatalf("Failed to open log file: %v", err)
	}
	defer lf.Close()
	defer logger.Init("LoggerExample", true, true, lf).Close()

	file, err := os.Open("config.json")
	defer file.Close()
	if err != nil {
		logger.Fatal("Error While opening config file", err)
	}
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err = decoder.Decode(&configuration)
	if err != nil {
		logger.Error(err)
	}
	redisClient = redis.NewClient(&redis.Options{
		Addr: configuration.Redis.Ip + ":" + strconv.Itoa(configuration.Redis.Port),
		DB:   0,
	})
	r := mux.NewRouter()
	r.HandleFunc("/dns", getDNS)
	r.HandleFunc("/ip", getIP)

	http.Handle("/", r)

	srv := &http.Server{
		Handler:      r,
		Addr:         configuration.Server,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	logger.Infoln("Server running at", configuration.Server)

	logger.Fatal(srv.ListenAndServe())
}

func getDNS(w http.ResponseWriter, r *http.Request) {
	var dnsServerips []string
	vars := r.URL.Query().Get("r")
	dnsip := redisClient.Keys(vars + ":*")
	for _, v := range dnsip.Val() {
		val := redisClient.Get(v)
		dnsServerips = append(dnsServerips, val.Val())
	}
	w.WriteHeader(http.StatusOK)
	js, _ := json.Marshal(dnsServerips)
	w.Write(js)
}

func getIP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	js, _ := json.Marshal(strings.Split(r.RemoteAddr, ":")[0])
	w.Write(js)
}

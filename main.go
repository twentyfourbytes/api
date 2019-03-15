package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis"

	"github.com/gorilla/mux"
)

var redisClient *redis.Client

const ip = "127.0.0.1"
const port = "6379"
const server = "127.0.0.1:8000"

type Configuration struct {
	Redis struct {
		Port int    `json:"port"`
		Ip   string `json:"ip"`
	} `json:"redis"`
}

func main() {
	file, _ := os.Open("config.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err := decoder.Decode(&configuration)
	if err != nil {
		fmt.Println(err)

	}
	fmt.Println(configuration)
	redisClient = redis.NewClient(&redis.Options{
		Addr: ip + ":" + port,
		DB:   0,
	})
	r := mux.NewRouter()
	r.HandleFunc("/dns", dns)
	http.Handle("/", r)

	srv := &http.Server{
		Handler:      r,
		Addr:         server,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	fmt.Println("Server running at", server)

	log.Fatal(srv.ListenAndServe())
}

func dns(w http.ResponseWriter, r *http.Request) {
	var dnsServerips []string
	vars := r.URL.Query().Get("r")
	fmt.Println(vars)
	dnsip := redisClient.Keys(vars + ":*")
	for _, v := range dnsip.Val() {
		dnsServerips = append(dnsServerips, redisClient.Get(v).String())
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "DNS: %v\n", dnsip)
}

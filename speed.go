package main

import (
	"crypto/rand"
	"encoding/json"
	"expvar"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/logger"
)

const Megabyte = 1024 * 1024
const Blocksize = 32 * 1024

var connections = expvar.NewInt("connections")
var downloads = expvar.NewInt("downloads")
var uploads = expvar.NewInt("uploads")
var downloadMegs = expvar.NewInt("downloadMegs")
var uploadMegs = expvar.NewInt("uploadMegs")

type Speed struct {
	Ip  string  `json:"ip"`
	MBs float64 `json:"megabytespersecond"` // Mega Bytes per second
	Mbs float64 `json:"megabitspersecond"`  // Mega Bits per second

	Duration  time.Duration `json:"duration"`
	Megabytes float64       `json:"fileseizeinmb"`
}

func download(w http.ResponseWriter, req *http.Request) {
	ms := req.FormValue("size")
	randID := req.FormValue("randID")

	if ms == "" {
		ms = "1"
	}
	m, err := strconv.ParseUint(ms, 10, 0)
	if err != nil {
		m = 1
	}
	w.Header().Set("Content-length", strconv.FormatUint(m*Megabyte, 10))
	w.Header().Set("Content-Type", "application/octet-stream")
	log.Printf("starting addr=%s megabytes=%d\n", extractIP(req.RemoteAddr), m)
	fileGenerator(w, req, m, randID)
}

func upload(w http.ResponseWriter, req *http.Request) {
	randID := req.FormValue("randID")

	log.Printf("upload starting addr=%s\n", extractIP(req.RemoteAddr))
	start := time.Now()
	status := "finished"

	written, err := io.Copy(ioutil.Discard, req.Body)
	if err != nil {
		status = "aborted"
		logger.Errorln("Error testing upload speed ", err)
	}

	duration := time.Since(start)
	megabytes := float64(written) / Megabyte
	MBs := megabytes / duration.Seconds()
	Mbs := MBs * 8
	message := fmt.Sprintf("upload %s addr=%s duration=%s megabytes=%.1f speed=%.1fMB/s\n", status, extractIP(req.RemoteAddr), duration, megabytes, MBs)
	logger.Infoln(message)
	s := &Speed{Ip: extractIP(req.RemoteAddr), Duration: duration, MBs: MBs, Mbs: Mbs, Megabytes: megabytes}
	js, _ := json.Marshal(s)
	redisClient.Set(randID, js, 0)
	w.Write(js)
}

func fileGenerator(w http.ResponseWriter, req *http.Request, m uint64, randID string) {
	addConnection()
	start := time.Now()
	status := "finished"
	written, err := io.Copy(w, LimitedRandomGen(m*Megabyte))

	if err != nil {
		status = "aborted"
	}
	duration := time.Since(start)
	megabytes := float64(written) / Megabyte
	mBs := megabytes / duration.Seconds()
	mbs := mBs * 8.0
	removeConnection(int64(megabytes), 0)
	s := &Speed{Ip: extractIP(req.RemoteAddr), Duration: duration, Megabytes: megabytes, MBs: mBs, Mbs: mbs}
	sjson, _ := json.Marshal(s)
	redisClient.Set(randID+strconv.FormatInt(int64(m), 10), sjson, 0)
	log.Printf("%s addr=%s duration=%s megabytes=%.1f speed=%.1fMB/s\n", status, extractIP(req.RemoteAddr), duration, megabytes, mbs)
}

func addConnection() {
	connections.Add(1)
}

func removeConnection(downmegs int64, upmegs int64) {
	connections.Add(-1)

	if downmegs != 0 {
		downloads.Add(1)
		downloadMegs.Add(downmegs)
	}
	if upmegs != 0 {
		uploads.Add(1)
		uploadMegs.Add(upmegs)
	}
}

func extractIP(addrPort string) string {
	lastColon := strings.LastIndex(addrPort, ":")
	return addrPort[0:lastColon]
}

type FileGen struct {
	buf []byte
}

func NewFileGen() *FileGen {
	randomBytes := make([]byte, Blocksize)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil
	}

	return &FileGen{
		buf: randomBytes,
	}
}

func (r *FileGen) Read(p []byte) (n int, err error) {
	n = len(p)
	toread := n
	if n > len(r.buf) {
		toread = len(r.buf)
	}

	copy(p[0:toread], r.buf[0:toread])
	return toread, nil
}

func LimitedRandomGen(n uint64) io.Reader {
	return io.LimitReader(NewFileGen(), int64(n))
}

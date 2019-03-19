package main

import (
	"crypto/rand"
	"expvar"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const Megabyte = 1024 * 1024
const Blocksize = 32 * 1024

var connections = expvar.NewInt("connections")
var downloads = expvar.NewInt("downloads")
var uploads = expvar.NewInt("uploads")
var downloadMegs = expvar.NewInt("downloadMegs")
var uploadMegs = expvar.NewInt("uploadMegs")

func download(w http.ResponseWriter, req *http.Request) {
	ms := req.FormValue("size")
	if ms == "" {
		ms = "1"
	}
	m, err := strconv.ParseUint(ms, 10, 0)
	if err != nil {
		m = 1
	}
	w.Header().Set("Content-length", strconv.FormatUint(m*Megabyte, 10))
	log.Printf("flood starting addr=%s megabytes=%d\n", extractIP(req.RemoteAddr), m)
	floodHelper(w, req, LimitedRandomGen(m*Megabyte))
}

func floodHelper(w http.ResponseWriter, req *http.Request, reader io.Reader) {
	addConnection()
	start := time.Now()
	status := "finished"
	written, err := io.Copy(w, reader)

	if err != nil {
		status = "aborted"
	}
	duration := time.Since(start)
	megabytes := float64(written) / Megabyte
	mbs := megabytes / duration.Seconds()
	removeConnection(int64(megabytes), 0)
	log.Printf("flood %s addr=%s duration=%s megabytes=%.1f speed=%.1fMB/s\n", status, extractIP(req.RemoteAddr), duration, megabytes, mbs)
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

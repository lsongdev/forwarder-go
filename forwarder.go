package main

import (
	"io"
	"log"
	"net"
	"sync/atomic"
)

type Forwarder struct {
	From            string
	To              string
	BytesUploaded   uint64
	BytesDownloaded uint64
	Connections     uint64
	listener        net.Listener
	quit            chan bool
}

func (f *Forwarder) Start() error {
	listener, err := net.Listen("tcp", f.From)
	if err != nil {
		return err
	}
	f.listener = listener
	f.quit = make(chan bool)
	go f.run()
	return nil
}

func (f *Forwarder) Stop() {
	close(f.quit)
	f.listener.Close()
}

func (f *Forwarder) run() {
	for {
		select {
		case <-f.quit:
			return
		default:
			conn, err := f.listener.Accept()
			if err != nil {
				log.Printf("Failed to accept connection: %v", err)
				continue
			}
			atomic.AddUint64(&f.Connections, 1)
			go f.forward(conn)
		}
	}
}

func (f *Forwarder) forward(src net.Conn) {
	dst, err := net.Dial("tcp", f.To)
	if err != nil {
		log.Printf("Unable to connect to destination %s: %v", f.To, err)
		src.Close()
		return
	}
	go f.copyData(dst, src, &f.BytesUploaded)   // src to dst (upload)
	go f.copyData(src, dst, &f.BytesDownloaded) // dst to src (download)
}

func (f *Forwarder) copyData(dst, src net.Conn, counter *uint64) {
	defer dst.Close()
	defer src.Close()
	n, _ := io.Copy(dst, src)
	atomic.AddUint64(counter, uint64(n))
}

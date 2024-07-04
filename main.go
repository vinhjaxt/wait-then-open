package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var bindAddr = flag.String(`bind`, `[::]:12312`, `TCP listen address port`)
var wd = flag.String(`wd`, `\\tsclient\RDP_Shared\`, `Working dir`)
var bufLen = flag.Int(`buf-len`, 512, `TCP buffer length`)
var timeout = flag.Duration(`timeout`, time.Second*30, `Timeout wait for file exists`)

func handle(conn net.Conn) {
	defer conn.Close()
	log.Println(`new conn`)

	buf := make([]byte, *bufLen)
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	n, err := conn.Read(buf)
	if err != nil {
		log.Println(`conn read:`, err)
		return
	}

	file_name := filepath.Base(strings.ReplaceAll(strings.Trim(string(buf[:n]), "\r\n\t \x00"), `/`, `\`))
	file_path := *wd + `\` + file_name
	if !strings.HasPrefix(filepath.Clean(file_path), *wd) {
		log.Println(`prefix filepath:`, file_path)
		return
	}
	log.Println(`filepath:`, file_path)
	conn.Write([]byte(file_path))
	conn.Write([]byte("\n"))

	t := time.NewTimer(*timeout)
	for {

		select {
		case <-t.C: // deadline
			log.Println(`deadline file not found`)
			return
		default: // continue
		}

		_, err = os.Stat(file_path)
		if err != nil {
			// file not exists
			time.Sleep(2 * time.Second)
			continue
		}

		// file exists
		log.Println(`file exists`)
		t.Stop()
		break
	}

	cmd := exec.Command("cmd.exe", "/C", "start", "", "/b", file_path)
	err = cmd.Run()
	if err != nil {
		log.Println(`exec:`, err)
		return
	}

	log.Println(`opened:`, file_path)
	conn.Write([]byte(`opened`))
	conn.Write([]byte("\n"))
}

func main() {
	flag.Parse()

	ln, err := net.Listen(`tcp`, *bindAddr)
	if err != nil {
		log.Panicln(`listen:`, err)
	}
	log.Println(`listen:`, ln.Addr().String())

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go handle(conn)
	}
}

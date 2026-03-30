//  Copyright(C) 2026 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2026-03-28

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	pid := os.Getpid()
	log.Println("ping.go running ...., pid=", pid)
	rd := bufio.NewReader(os.Stdin)
	for {
		line, err := rd.ReadString('\n')
		if err != nil {
			log.Fatalf("ping.go pid= %v  exit with error: %v", pid, err)
		}
		line = strings.TrimSpace(line)
		log.Printf("ping.go pid=%d read= %q\n", pid, line)
		fmt.Fprintf(os.Stdout, "Ok: %s\n", line)
	}
	log.Println("ping.go exiting ...., pid=", pid)
}

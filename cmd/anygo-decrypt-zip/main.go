//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-01-02

package main

import (
	"bytes"
	"crypto/md5"
	"flag"
	"log"
	"os"

	"github.com/xanygo/anygo/xcodec"
)

var token = flag.String("token", "anygo-3000", "token for encryption")

func main() {
	flag.Parse()
	files := flag.Args()
	if len(files) == 0 {
		log.Fatal("no files to decrypt")
	}
	dz := &xcodec.AesOFB{
		Key: *token,
	}
	for _, file := range files {
		decodeFile(dz, file)
	}
}

func decodeFile(dz *xcodec.AesOFB, file string) {
	content, err := os.ReadFile(file)
	assert(err, "ReadFile "+file)
	if len(content) < 32 {
		log.Fatal("file too short", len(content))
	}
	etContent := content[:len(content)-16]
	xm := md5.New()
	xm.Write(etContent)
	xm.Write(dz.ID())
	sign := xm.Sum(nil)
	expect := content[len(content)-16:]
	if !bytes.Equal(expect, sign) {
		log.Fatalln("invalid signature")
	}

	pt, err := dz.Decrypt(etContent)
	assert(err, "Decrypt "+file)

	outFp := file + ".zip"
	err = os.WriteFile(outFp, pt, 0644)
	assert(err, "WriteFile "+outFp)
	log.Printf("Decrypted file %s to %s", file, outFp)
}

func assert(err error, msg string) {
	if err != nil {
		log.Fatalln(msg, err)
	}
}

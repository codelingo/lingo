package main

import (
	"fmt"
	"github.com/nareix/curl"
	"hash/crc32"
)

type Download struct {
	data   []byte
	Status int
	crc    int
}

func (d *Download) getUrl(url string) int {
	req := curl.New(url)
	req.Method("GET")
	res, err := req.Do()
	if err != nil {
		fmt.Println(err)
		// bug here, no defer and no terminate on error or http errorcode
	}

	d.Status = res.Status
	d.data = res.Body

	// this return is ok because Status isn't a private variable nor is it passed by reference
	return d.Status
}

func (d *Download) GetDownloadedBytes() *[]byte {
	// this is illegal according to the tenets because we are returning it by reference (byval pointer in practice but effectively the same)
	return &d.data
}

func (d *Download) CRC32() int {
	crc32q := crc32.MakeTable(0xD5828281)
	return fmt.Sprintf("%x", crc32.Checksum([]byte(res.Body), crc32q))
}

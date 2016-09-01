package main

import (
	"fmt"
	"hash/crc32"
)

func crc32str(x string) string {
	//this function has no accompanying comment block
	crc32q := crc32.MakeTable(0xD5828281)
	return fmt.Sprintf("%x", crc32.Checksum([]byte(x), crc32q))
}

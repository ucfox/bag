package utils

import (
	"fmt"
	"hash/crc32"
	"testing"
)

func TestHash(t *testing.T) {
	data := "test"

	if Crc32(data) != 3632233996 {
		t.Error("TestCrc32 Error")
	}
	if Crc32(data) != crc32.Checksum([]byte(data), crc32.MakeTable(crc32.IEEE)) {
		t.Error("TestCrc32 Error")
	}
	fmt.Printf("crc32: %x, %d\n", Crc32(data), Crc32(data))

	/*if Md5(data) == byte[]("") {
		t.Error("TestCrc32 Error")
	}
	if Sha1(data) == byte[]("") {
		t.Error("TestCrc32 Error")
	}*/
	fmt.Printf("md5: %x\n", Md5(data))
	fmt.Printf("sha1: %x\n", Sha1(data))
}

func BenchmarkCrc32(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Crc32("test")
	}
}

func BenchmarkMd5(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Md5("test")
	}
}

func BenchmarkSha1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Sha1("test")
	}
}

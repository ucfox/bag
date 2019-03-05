package utils
import (
	"hash/crc32"
	"crypto/md5"
	"crypto/sha1"
)

func Crc32(str string) uint32 {
	return crc32.ChecksumIEEE([]byte(str))
}

func Md5(str string) [16]byte {
	return md5.Sum([]byte(str))
}

func Sha1(str string) [20]byte {
	return sha1.Sum([]byte(str))
}

package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var (
	binPath  = flag.String("bin_path", "./tests/min.sint64", "path to the desired bin file to encode/decode")
	dataType = flag.String("type", "", "data type to encode/decode (uint64, sint64)")
)

const (
	PAYLOAD_BITS = 7
	PAYLOAD_MASK = 0x7f
	CONT_MASK    = 0x80
)

func decodeU64(bs []byte) uint64 {
	decoded := uint64(0)

	for i, b := range bs {
		decoded |= uint64(b&PAYLOAD_MASK) << (i * PAYLOAD_BITS)
	}

	return decoded
}

func decodeSI64(bs []byte) int64 {
	decoded := decodeU64(bs)
	return int64((decoded >> 1) ^ -(decoded & 0x1))
}

func decodeStream[V uint64 | int64](bs []byte) []V {
	decoded := make([]V, 0)

	var appF func([]byte) []V
	switch any(decoded).(type) {
	case []uint64:
		{
			appF = func(curBs []byte) []V {
				return append(decoded, V(decodeU64(curBs)))
			}
		}
	case []int64:
		{
			appF = func(curBs []byte) []V {
				return append(decoded, V(decodeSI64(curBs)))
			}
		}
	}

	curBs := make([]byte, 0)
	for _, b := range bs {
		curBs = append(curBs, b)
		if b&CONT_MASK == 0 {
			decoded = appF(curBs)
			curBs = make([]byte, 0)
		}
	}

	return decoded
}

func encodeU64(val uint64) []byte {
	bs := make([]byte, 0, 10)

	for val != 0 {
		cur := byte(val & PAYLOAD_MASK)
		val >>= PAYLOAD_BITS
		if val != 0 {
			cur |= CONT_MASK
		}
		bs = append(bs, cur)
	}

	if len(bs) == 0 {
		bs = append(bs, 0)
	}

	return bs
}

func encodeSI64(val int64) []byte {
	return encodeU64(uint64(val+val) ^ -((uint64(val) & (1 << 63)) >> 63))
}

func encodeStream[V uint64 | int64](bs []byte) []byte {
	ints := make([]V, 0, len(bs)/8)
	for i := 0; i < len(bs); i += 8 {
		cur := binary.BigEndian.Uint64(bs[i : i+8])
		ints = append(ints, V(cur))
	}
	fmt.Fprintf(os.Stderr, "input stream: %#02v == %d\n", bs, ints)

	stream := make([]byte, 0)
	var appF func(V) []byte
	switch any(ints).(type) {
	case []uint64:
		{
			appF = func(v V) []byte {
				return append(stream, encodeU64(uint64(v))...)
			}
		}
	case []int64:
		{
			appF = func(v V) []byte {
				return append(stream, encodeSI64(int64(v))...)
			}
		}
	}
	for _, v := range ints {
		stream = appF(v)
	}
	return stream
}

func check(err error) {
	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(1)
	}
}

func main() {
	flag.Parse()

	file, err := os.Open(*binPath)
	check(err)
	bs, err := io.ReadAll(file)
	check(err)

	if *dataType == "" {
		*dataType = filepath.Ext(*binPath)[1:]
	}

	var encF func([]byte) []byte
	var decF func([]byte) interface{}
	switch *dataType {
	case "uint64":
		{
			encF = encodeStream[uint64]
			decF = func(bs []byte) interface{} {
				return decodeStream[uint64](bs)
			}
		}
	case "sint64":
		{
			encF = encodeStream[int64]
			decF = func(bs []byte) interface{} {
				return decodeStream[int64](bs)
			}
		}
	default:
		{
			fmt.Printf("ERROR: invalid value: %s for command line argument 'dataType'\n", *dataType)
			os.Exit(1)
		}
	}

	encodedStream := encF(bs)
	fmt.Fprintf(os.Stderr, "encoded stream: %#02v\n", encodedStream)
	fmt.Fprintf(os.Stderr, "decoded stream: %d\n", decF(encodedStream))
}

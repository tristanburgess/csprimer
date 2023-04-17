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
	cmd      = flag.String("cmd", "decode", "command to run (encode, decode, decodeStream)")
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

func decodeU64Stream(bs []byte) []uint64 {
	decoded := make([]uint64, 0)

	curBs := make([]byte, 0)
	for _, b := range bs {
		curBs = append(curBs, b)
		if b&CONT_MASK == 0 {
			decoded = append(decoded, decodeU64(curBs))
			curBs = make([]byte, 0)
		}
	}

	return decoded
}

func decodeSI64(bs []byte) int64 {
	decoded := decodeU64(bs)
	return int64((decoded >> 1) ^ -(decoded & 0x1))
}

func decodeSI64Stream(bs []byte) []int64 {
	decoded := make([]int64, 0)

	curBs := make([]byte, 0)
	for _, b := range bs {
		curBs = append(curBs, b)
		if b&CONT_MASK == 0 {
			decoded = append(decoded, decodeSI64(curBs))
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

	if *dataType != "uint64" && *dataType != "sint64" {
		fmt.Printf("ERROR: invalid value: %s for command line argument 'dataType'\n", *dataType)
		os.Exit(1)
	}

	switch *cmd {
	case "encode":
		{
			if *dataType == "uint64" {
				val := binary.BigEndian.Uint64(bs)
				fmt.Fprintf(os.Stderr, "input: %#02v == %d\n", bs, val)
				fmt.Fprintf(os.Stderr, "encoded: %#02v\n", encodeU64(val))
			} else if *dataType == "sint64" {
				val := int64(binary.BigEndian.Uint64(bs))
				fmt.Fprintf(os.Stderr, "input: %#02v == %d\n", bs, val)
				fmt.Fprintf(os.Stderr, "encoded: %#02v\n", encodeSI64(val))
			}
		}
	case "decode":
		{
			if *dataType == "uint64" {
				val := binary.BigEndian.Uint64(bs)
				fmt.Fprintf(os.Stderr, "input: %#02v == %d\n", bs, val)

				encoded := encodeU64(val)
				fmt.Fprintf(os.Stderr, "encoded: %#02v\n", encoded)
				fmt.Fprintf(os.Stderr, "decoded: %d\n", decodeU64(encoded))
			} else if *dataType == "sint64" {
				val := int64(binary.BigEndian.Uint64(bs))
				fmt.Fprintf(os.Stderr, "input: %#02v == %d\n", bs, val)

				encoded := encodeSI64(val)
				fmt.Fprintf(os.Stderr, "encoded: %#02v\n", encoded)
				fmt.Fprintf(os.Stderr, "decoded: %d\n", decodeSI64(encoded))
			}
		}
	case "decodeStream":
		{
			if *dataType == "uint64" {
				uints := make([]uint64, 0, len(bs)/8)
				for i := 0; i < len(bs); i += 8 {
					uints = append(uints, binary.BigEndian.Uint64(bs[i:i+8]))
				}
				fmt.Fprintf(os.Stderr, "input: %#02v == %d\n", bs, uints)
				stream := make([]byte, 0)
				for _, v := range uints {
					stream = append(stream, encodeU64(v)...)
				}
				fmt.Fprintf(os.Stderr, "encoded: %#02v\n", stream)
				fmt.Fprintf(os.Stderr, "decoded: %d\n", decodeU64Stream(stream))
			} else if *dataType == "sint64" {
				ints := make([]int64, 0, len(bs)/8)
				for i := 0; i < len(bs); i += 8 {
					ints = append(ints, int64(binary.BigEndian.Uint64(bs[i:i+8])))
				}
				fmt.Fprintf(os.Stderr, "input: %#02v == %d\n", bs, ints)
				stream := make([]byte, 0)
				for _, v := range ints {
					stream = append(stream, encodeSI64(v)...)
				}
				fmt.Fprintf(os.Stderr, "encoded: %#02v\n", stream)
				fmt.Fprintf(os.Stderr, "decoded: %d\n", decodeSI64Stream(stream))
			}
		}
	default:
		{
			fmt.Printf("ERROR: invalid value: %s for command line argument 'cmd'\n", *cmd)
			os.Exit(1)
		}
	}
}

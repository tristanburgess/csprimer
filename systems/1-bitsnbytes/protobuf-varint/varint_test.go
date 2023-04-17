package main

import (
	"math"
	"reflect"
	"testing"
)

func TestDecodeSI64Stream(t *testing.T) {
	ints := make([]int64, 0, 4*(1<<20))
	for i := int64(0); i < 1<<20; i++ {
		ints = append(ints, i)
		ints = append(ints, -i)
		ints = append(ints, math.MaxInt64-i)
		ints = append(ints, math.MinInt64+i)
	}
	stream := make([]byte, 0)
	for _, v := range ints {
		stream = append(stream, encodeSI64(v)...)
	}
	decoded := decodeStream[int64](stream)
	if !reflect.DeepEqual(decoded, ints) {
		t.Fatalf("decodeSI64Stream() == %02v, which did not match sint64 stream", decoded)
	}
}

func TestDecodeU64Stream(t *testing.T) {
	uints := make([]uint64, 0, 2*(1<<20))
	for i := uint64(0); i < 1<<20; i++ {
		uints = append(uints, i)
		uints = append(uints, math.MaxUint64-i)
	}
	stream := make([]byte, 0)
	for _, v := range uints {
		stream = append(stream, encodeU64(v)...)
	}
	decoded := decodeStream[uint64](stream)
	if !reflect.DeepEqual(decoded, uints) {
		t.Fatalf("decodeU64Stream() == %02v, which did not match uint64 stream", decoded)
	}
}

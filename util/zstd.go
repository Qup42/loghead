package util

import (
	"github.com/klauspost/compress/zstd"
	"sync"
	"tailscale.com/smallzstd"
)

func ZstdDecode(in []byte) []byte {
	decoder, ok := zstdDecoderPool.Get().(*zstd.Decoder)
	if !ok {
		panic("invalid type in sync pool")
	}
	out, _ := decoder.DecodeAll(in, nil)
	_ = decoder.Reset(nil)
	zstdDecoderPool.Put(decoder)

	return out
}

var zstdDecoderPool = &sync.Pool{
	New: func() any {
		encoder, err := smallzstd.NewDecoder(
			nil)
		if err != nil {
			panic(err)
		}

		return encoder
	},
}

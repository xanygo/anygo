package xenc

type RadixCodec interface {
	EncodeInt(n int64) (string, error)
	DecodeInt(str string) (int64, error)
}

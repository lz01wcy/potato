package net

type ICodec interface {
	Decode([]byte) (interface{}, error)
	Encode(interface{}) ([]byte, error)
}

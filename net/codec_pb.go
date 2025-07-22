package net

type PbCodec struct {
}

func (c *PbCodec) Encode(v interface{}) (msgBytes []byte, err error) {
	return
}

func (c *PbCodec) Decode(data []byte) (msg interface{}, err error) {
	return
}

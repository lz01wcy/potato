package net

import (
	"encoding/binary"
	"errors"
	"io"
)

const (
	lenSize     = 4           // 包体大小字段 大小为id+body总共长度
	maxPackSize = 1024 * 1024 //消息最大长度
)

var (
	ErrMaxPacket = errors.New("packet over size")
	ErrMinPacket = errors.New("packet short size")
)

// 接收Length-Value格式的封包流程 返回包中的V
func ReadLVPacket(reader io.Reader) (v []byte, err error) {

	// Size为uint32，占4字节
	var sizeBuffer = make([]byte, lenSize)

	// 持续读取Size直到读到为止
	_, err = io.ReadFull(reader, sizeBuffer)

	// 发生错误时返回
	if err != nil {
		return
	}

	if len(sizeBuffer) < lenSize {
		return nil, ErrMinPacket
	}

	// 用小端格式读取Size
	bodyLen := binary.LittleEndian.Uint32(sizeBuffer)

	if int(bodyLen) > maxPackSize {
		return nil, ErrMaxPacket
	}

	// 分配包体大小
	v = make([]byte, bodyLen)

	// 读取包体数据
	_, err = io.ReadFull(reader, v)

	return
}

// 发送Length-Value格式的封包
func WriteLVPacket(writer io.Writer, pkt []byte) error {
	// 将数据写入Socket
	total := len(pkt)

	for pos := 0; pos < total; {

		n, err := writer.Write(pkt[pos:])

		if err != nil {
			return err
		}

		pos += n
	}

	return nil
}

// 解包
func UnpackPacket(codec ICodec, msg []byte) (interface{}, error) {
	// 将字节数组和消息ID用户解出消息
	pack, err := codec.Decode(msg)
	if err != nil {
		return nil, err
	}
	return pack, nil
}

// 封包
func PackPacket(codec ICodec, msg interface{}) ([]byte, error) {
	msgData, err := codec.Encode(msg)
	if err != nil {
		return nil, err
	}
	pkt := make([]byte, lenSize+len(msgData))

	// Length
	binary.LittleEndian.PutUint32(pkt, uint32(len(msgData)))

	// Value
	copy(pkt[lenSize:], msgData)

	return pkt, nil
}

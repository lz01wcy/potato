package pb

import (
	"github.com/murang/potato/log"
	"os"
	"reflect"
)

var (
	id2Type = make(map[uint32]reflect.Type) // msgId -> type
	type2Id = make(map[reflect.Type]uint32) // type -> msgId
)

func RegisterMsg(msgId uint32, msgType reflect.Type) {
	if _, ok := type2Id[msgType]; ok {
		log.Sugar.Errorf("RegisterMsgMateType2Id err, msg repeat : %s", msgType.String())
		os.Exit(2)
	}
	type2Id[msgType] = msgId
	if _, ok := id2Type[msgId]; ok {
		log.Sugar.Errorf("RegisterC2SMsgMate err, msg repeat : %d", msgId)
		os.Exit(2)
	}
	id2Type[msgId] = msgType
}

func GetIdByType(t reflect.Type) uint32 {
	id, ok := type2Id[t]
	if !ok {
		return 0
	}
	return id
}

func GetTypeById(id uint32) reflect.Type {
	t, ok := id2Type[id]
	if !ok {
		return nil
	}
	return t
}

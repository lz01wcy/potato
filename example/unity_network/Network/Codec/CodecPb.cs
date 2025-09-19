using System;
using Google.Protobuf;
using UnityEngine;

namespace GameLogic{
public class CodecPb : ICodec{

    public byte[] Encode(object msg)
    {
        // 先通过消息类型找到id
        int id = MsgMeta.GetIdByType(msg.GetType());
        byte[] data = ((IMessage)msg).ToByteArray();
        
        byte[] buf = new byte[data.Length + 4];
        var idBytes = BitConverter.GetBytes(id);
        if (BitConverter.IsLittleEndian)
        {
            Array.Reverse(idBytes);
        }
        Buffer.BlockCopy(idBytes, 0, buf, 0, 4);
        Buffer.BlockCopy(data, 0, buf, 4, data.Length);
        return buf;
    }

    public object Decode(byte[] data)
    {
        var idBytes = data[..4];
        if (BitConverter.IsLittleEndian)
        {
            Array.Reverse(idBytes);
        }
        int id = BitConverter.ToInt32(idBytes, 0);
        var msgType = MsgMeta.GetTypeById(id);
        if (msgType == null)
        {
            Debug.LogError("can not get msg type by id");
            return null;
        }
        var msg = (IMessage)Activator.CreateInstance(msgType);
        msg = msg.Descriptor.Parser.ParseFrom(data, 4, data.Length - 4);

        return msg;
    }
}
}

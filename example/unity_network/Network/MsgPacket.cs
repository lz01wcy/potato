using System;
using System.IO;
using Cysharp.Threading.Tasks;

namespace GameLogic
{
    public class MsgPack
    {
        const int BodyLenSize = 4;
        const int MaxPackSize = 1024 * 1024;

        public static async UniTask<byte[]> ReceivePacket(Stream stream)
        {
            var lenBytes = await ReadFull(stream, BodyLenSize);
            if (BitConverter.IsLittleEndian)
            {
                Array.Reverse(lenBytes);
            }
            int len = BitConverter.ToInt32(lenBytes, 0);
            if (len > MaxPackSize)
            {
                Log.Error("RecvLvPacket over size : " + len);
            }

            var msgBytes = await ReadFull(stream, len);
            return msgBytes;
        }

        public static byte[] PackMsgData(byte[] msgData)
        {

            byte[] packData = new byte[BodyLenSize + msgData.Length];
            var lenBytes = BitConverter.GetBytes(msgData.Length);
            if (BitConverter.IsLittleEndian)
            {
                Array.Reverse(lenBytes);
            }
            Array.Copy(lenBytes, packData, BodyLenSize);
            Array.Copy(msgData, 0, packData, BodyLenSize, msgData.Length);
            return packData;
        }

        /// <summary>
        /// 阻塞直到读满
        /// </summary>
        /// <param name="stream"></param>
        /// <param name="len"></param>
        private static async UniTask<byte[]> ReadFull(Stream stream, int len)
        {
            int offset = 0;
            byte[] buf = new byte[len];
            while (offset < len)
            {
                offset += await stream.ReadAsync(buf, offset, len - offset);
            }

            return buf;
        }
    }
}
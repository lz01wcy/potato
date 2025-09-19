using System;
using UnityWebSocket;

namespace GameLogic
{
    public class WsConn : IConn
    {
        private ConnOption _option;
        private WebSocket _socket;
        
        public WsConn(ConnOption option)
        {
            _option = option;
            if (_option.Domain != null)
            {
                _socket = new WebSocket(option.Domain);
            }
            else
            {
                _socket = new WebSocket($"ws://{_option.Host}:{_option.Port}");
            }
            

            _socket.OnOpen += (sender, args) => OnConnected?.Invoke(this);
            _socket.OnClose += (sender, args) => OnDisconnected?.Invoke(this);
            _socket.OnError += (sender, args) => OnException?.Invoke(this, new Exception(args.Message));
            _socket.OnMessage += (sender, args) =>
            {
                // 前面4个字节是消息长度 ws中已经自动处理了 需要跳过
                var msgData = new byte[args.RawData.Length - 4];
                Array.Copy(args.RawData, 4, msgData, 0, msgData.Length);
                object msg = _option.Codec.Decode(msgData);
                OnMessage?.Invoke(this, msg);
            };
        }

        public void Dispose()
        {
            _socket.CloseAsync();
        }

        public void Connect()
        {
            _socket.ConnectAsync();
        }

        public void Send(object message)
        {
            byte[] data = _option.Codec.Encode(message);
            _socket.SendAsync(MsgPack.PackMsgData(data));
        }

        public void Disconnect()
        {
            _socket.CloseAsync();
        }

        public event Action<IConn, Exception> OnException;
        public event Action<IConn> OnConnected;
        public event Action<IConn> OnDisconnected;
        public event Action<IConn, object> OnMessage;
    }
}

using System;
using System.IO;
using System.Net.Sockets;
using Cysharp.Threading.Tasks;

namespace GameLogic
{
    public struct ConnOption
    {
        public string Domain;
        public string Host;
        public int Port;
        public ICodec Codec;
        public float ConnectTimeout;
        public float Timeout;
    }
    
    public interface IConn : IDisposable
    {
        void Connect();
        void Send(object message);
        void Disconnect();

        event Action<IConn, Exception> OnException; // 出错
        event Action<IConn> OnConnected; // 连接成功事件
        event Action<IConn> OnDisconnected; // 断开连接事件
        event Action<IConn, object> OnMessage; // 收到消息事件
    }
    
    public abstract class Conn : IConn
    {
        protected ConnOption Option;
        
        protected Socket _socket;
        protected Stream _stream;
        protected bool _running;
        
        public event Action<IConn, Exception> OnException;
        public event Action<IConn> OnConnected;
        public event Action<IConn> OnDisconnected;
        public event Action<IConn, object> OnMessage;
        
        public async void Connect()
        {
            if (_socket == null)
            {
                throw new Exception("Socket is null");
            }
            
            try
            {
                if (_socket.ProtocolType == ProtocolType.Tcp)
                {
                    await _socket.ConnectAsync(Option.Host, Option.Port)
                        .AsUniTask()
                        .Timeout(TimeSpan.FromSeconds(Option.ConnectTimeout));
                    _stream = new NetworkStream(_socket, true);
                }else if (_socket.ProtocolType == ProtocolType.Udp)
                {
                    _socket.Connect(Option.Host, Option.Port); // udp不存在握手 直接默认连接成功
                    _stream = new KCPStream(_socket, true);
                }
                
                _running = true;
                OnConnected?.Invoke(this);
            }
            catch (Exception e)
            {
                OnException?.Invoke(this, e);
                throw;
            }
            
            ReceiveLoop().Forget();
        }
        
        public async void Send(object message)
        {
            if (_stream == null) return;
            byte[] data = Option.Codec.Encode(message);
            var pack = MsgPack.PackMsgData(data);
            try
            {
                await _stream.WriteAsync(pack, 0, pack.Length);
                await _stream.FlushAsync();
            }
            catch (Exception e)
            {
                OnException?.Invoke(this, e);
                Disconnect();
            }
        }

        public void Disconnect()
        {
            if (!_running) return;
            _running = false;

            _stream?.Close();
            _socket?.Close();
 
            OnDisconnected?.Invoke(this);
        }

        private async UniTaskVoid ReceiveLoop()
        {
            try
            {
                while (_running)
                {
                    var result = await MsgPack.ReceivePacket(_stream)
                        .TimeoutWithoutException(TimeSpan.FromSeconds(Option.Timeout));
                    if (result.IsTimeout)
                    {
                        throw new Exception("Receive timeout");
                    }
                    
                    object msg = Option.Codec.Decode(result.Result);
                    await UniTask.SwitchToMainThread();
                    OnMessage?.Invoke(this, msg);
                }
            }
            catch (Exception e)
            {
                OnException?.Invoke(this, e);
                Disconnect();
            }
        }
        
        public void Dispose()
        {
            Disconnect();
        }
    }
}

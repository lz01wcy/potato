using System.Net.Sockets;

namespace GameLogic
{
    public class TcpConn : Conn
    {
        public TcpConn(ConnOption option)
        {
            Option = option;
            
            _socket = new Socket(AddressFamily.InterNetwork, SocketType.Stream, ProtocolType.Tcp);
        }
    }
}

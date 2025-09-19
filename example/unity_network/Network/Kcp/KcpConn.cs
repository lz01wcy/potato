using System.Net.Sockets;

namespace GameLogic
{
    public class KcpConn : Conn
    {
        public KcpConn(ConnOption option)
        {
            Option = option;
            
            _socket = new Socket(AddressFamily.InterNetwork, SocketType.Dgram, ProtocolType.Udp);
        }
    }
}
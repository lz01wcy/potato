using Pb;

namespace GameLogic
{
    public class Useage
    {
        void TryConnect()
        {
            var msg = new Pb.C2S_Hello
            {
                Name = "niceman",
                User = new UserInfo
                {
                    Id = 666
                },
            };
            MsgDispatcher.Instance.RegisterMsgReceiver<Pb.C2S_Hello>(OnC2SHello);
            var codec = new CodecPb();
            var option = new ConnOption
            {
                Host = "127.0.0.1",
                Port = 10086,
                Codec = codec,
                ConnectTimeout = 5,
                Timeout = 30
            };
            var conn = new WsConn(option); // or new TcpConn/KcpConn
            conn.OnConnected += conn =>
            {
                Log.Debug("Connected");
                conn.Send(msg);
            };
            conn.OnMessage += (conn, o) =>
            {
                MsgDispatcher.Instance.DispatchMsg(o);
            };
            conn.OnException += (conn1, exception) =>
            {
                Log.Error(exception.Message);
            };
            conn.Connect();
        }

        void OnC2SHello(C2S_Hello msg)
        {
            Log.Debug("OnC2SHello =====> : " + msg);
        }
    }
}
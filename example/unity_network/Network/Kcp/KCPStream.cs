using System;
using System.IO;
using System.Net;
using System.Net.Sockets;

namespace GameLogic
{
    public class KCPStream : Stream
    {
        private Socket mSocket = null;
        private bool mOwnsSocket = false;
        private KCP mKCP = null;
        private bool mClosed = false;

        private ByteBuffer mRecvBuffer = ByteBuffer.Allocate(1024 * 32);
        private UInt32 mNextUpdateTime = 0;

        public bool IsConnected { get { return mSocket != null && mSocket.Connected; } }
        public bool WriteDelay { get; set; }
        public bool AckNoDelay { get; set; }

        public IPEndPoint RemoteAddress { get; private set; }
        public IPEndPoint LocalAddress { get; private set; }

        private DateTime startDt = DateTime.Now;
        const int logmask = KCP.IKCP_LOG_IN_ACK | KCP.IKCP_LOG_OUT_ACK | KCP.IKCP_LOG_IN_DATA | KCP.IKCP_LOG_OUT_DATA;

        public KCPStream(Socket s, bool ownsSocket)
        {
            mSocket = s;
            mOwnsSocket = ownsSocket;
            RemoteAddress = (IPEndPoint)mSocket.RemoteEndPoint;
            LocalAddress = (IPEndPoint)mSocket.LocalEndPoint;
            mKCP = new KCP((uint)(new Random().Next(1, Int32.MaxValue)), rawSend);
            // normal:  0, 40, 2, 1
            // fast:    0, 30, 2, 1
            // fast2:   1, 20, 2, 1
            // fast3:   1, 10, 2, 1
            mKCP.NoDelay(1, 10, 2, 1);
            mKCP.SetStreamMode(true);

            // Log
            //mKCP.SetLogger(Log);
            //mKCP.SetLogMask(logmask);

            mRecvBuffer.Clear();
        }
        
        public override void Close()
        {
            if (mSocket != null) {
                mSocket.Close();
                mSocket = null;
                mRecvBuffer.Clear();
            }

            mClosed = true;
        }
        
        private void rawSend(byte[] data, int length)
        {
            if (mSocket != null) {
                mSocket.Send(data, length, SocketFlags.None);
            }
        }
        
        public int Send(byte[] data, int index, int length)
        {
            if (mSocket == null)
                return -1;

            var waitsnd = mKCP.WaitSnd;
            if (waitsnd < mKCP.SndWnd && waitsnd < mKCP.RmtWnd) {

                var sendBytes = 0;
                do {
                    var n = Math.Min((int)mKCP.Mss, length - sendBytes);
                    mKCP.Send(data, index + sendBytes, n);
                    sendBytes += n;
                } while (sendBytes < length);

                waitsnd = mKCP.WaitSnd;
                if (waitsnd >= mKCP.SndWnd || waitsnd >= mKCP.RmtWnd || !WriteDelay) {
                    mKCP.Flush(false);
                }

                return length;
            }

            return 0;
        }

        public int Recv(byte[] data, int index, int length)
        {
            // 上次剩下的部分
            if (mRecvBuffer.ReadableBytes > 0) {
                var recvBytes = Math.Min(mRecvBuffer.ReadableBytes, length);
                Buffer.BlockCopy(mRecvBuffer.RawBuffer, mRecvBuffer.ReaderIndex, data, index, recvBytes);
                mRecvBuffer.ReaderIndex += recvBytes;
                // 读完重置读写指针
                if (mRecvBuffer.ReaderIndex == mRecvBuffer.WriterIndex) {
                    mRecvBuffer.Clear();
                }
                return recvBytes;
            }

            if (mSocket == null)
                return -1;

            if (!mSocket.Poll(0, SelectMode.SelectRead)) {
                return 0;
            }

            var rn = 0;
            try {
                rn = mSocket.Receive(mRecvBuffer.RawBuffer, mRecvBuffer.WriterIndex, mRecvBuffer.WritableBytes, SocketFlags.None);
            } catch(Exception ex) {
                Console.WriteLine(ex);
                rn = -1;
            }

            if (rn <= 0) {
                return rn;
            }
            mRecvBuffer.WriterIndex += rn;

            var inputN = mKCP.Input(mRecvBuffer.RawBuffer, mRecvBuffer.ReaderIndex, mRecvBuffer.ReadableBytes, true, AckNoDelay);
            if (inputN < 0) {
                mRecvBuffer.Clear();
                return inputN;
            }
            mRecvBuffer.Clear();

            // 读完所有完整的消息
            for (;;) {
                var size = mKCP.PeekSize();
                if (size < 0) break; //读不到了

                mRecvBuffer.EnsureWritableBytes(size);

                var n = mKCP.Recv(mRecvBuffer.RawBuffer, mRecvBuffer.WriterIndex, size);
                if (n > 0) mRecvBuffer.WriterIndex += n;
            }

            // 有数据待接收
            if (mRecvBuffer.ReadableBytes > 0) {
                return Recv(data, index, length);
            }

            return 0;
        }

        public void Update()
        {
            if (mSocket == null)
                return;

            if (0 == mNextUpdateTime || mKCP.CurrentMS >= mNextUpdateTime)
            {
                mKCP.Update();
                mNextUpdateTime = mKCP.Check();
            }
        }

        public void Log(string str)
        {
            DateTime now = DateTime.Now;
            int t = (int)(now - startDt).TotalMilliseconds;
            Console.WriteLine($"[{t.ToString().PadLeft(10, ' ')}] {str}");
        }
        
        public override void Flush()
        {
            mKCP.Flush(false);
        }
    
        public override int Read(byte[] buffer, int offset, int count)
        {
            object k = mKCP;
            int got = 0;
            lock (k)
            {
                Update();
                got = Recv(buffer, offset, count);
            }

            return got;
        }
        public override void Write(byte[] buffer, int offset, int count)
        {
            object k = mKCP;
            lock (k)
            {
                Send(buffer, offset, count);
            }
        }
    
        public override long Seek(long offset, SeekOrigin origin)
        {
            throw new System.NotSupportedException();
        }
    
        public override void SetLength(long value)
        {
            throw new System.NotSupportedException();
        }



        public override bool CanRead
        {
            get { return !mClosed; }
        }

        public override bool CanSeek
        {
            get { return false; }
        }

        public override bool CanWrite
        {
            get{ return !mClosed; }
        }
        public override long Length { get; }
        public override long Position { get; set; }
    }

}

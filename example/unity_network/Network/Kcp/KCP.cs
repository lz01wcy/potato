using System;
using System.Collections.Generic;

namespace GameLogic
{
    public class KCP
    {
        public const int IKCP_RTO_NDL = 30;  // 最小重传时间-无延迟
        public const int IKCP_RTO_MIN = 100; // 最小重传时间-正常
        public const int IKCP_RTO_DEF = 200;    
        public const int IKCP_RTO_MAX = 60000;
        public const int IKCP_CMD_PUSH = 81; // 推送命令
        public const int IKCP_CMD_ACK = 82; // 确认命令
        public const int IKCP_CMD_WASK = 83; // 接受窗口询问大小命令
        public const int IKCP_CMD_WINS = 84; // 接受涌口大小告知命令
        public const int IKCP_ASK_SEND = 1;  // 请求远端窗口大小
        public const int IKCP_ASK_TELL = 2;  // 请求告知远端窗口大小
        public const int IKCP_WND_SND = 32;
        public const int IKCP_WND_RCV = 32;
        public const int IKCP_MTU_DEF = 1400;
        public const int IKCP_ACK_FAST = 3;
        public const int IKCP_INTERVAL = 100;
        public const int IKCP_OVERHEAD = 24;    //数据包头长度
        public const int IKCP_DEADLINK = 20;
        public const int IKCP_THRESH_INIT = 2;
        public const int IKCP_THRESH_MIN = 2;
        public const int IKCP_PROBE_INIT = 7000;   // 7 secs to probe window size
        public const int IKCP_PROBE_LIMIT = 120000; // up to 120 secs to probe window
        public const int IKCP_SN_OFFSET = 12;


        // encode 8 bits unsigned int
        public static int ikcp_encode8u(byte[] p, int offset, byte c)
        {
            p[0 + offset] = c;
            return 1;
        }

        // decode 8 bits unsigned int
        public static int ikcp_decode8u(byte[] p, int offset, ref byte c)
        {
            c = p[0 + offset];
            return 1;
        }

        /* encode 16 bits unsigned int (lsb) */
        public static int ikcp_encode16u(byte[] p, int offset, UInt16 w)
        {
            p[0 + offset] = (byte)(w >> 0);
            p[1 + offset] = (byte)(w >> 8);
            return 2;
        }

        /* decode 16 bits unsigned int (lsb) */
        public static int ikcp_decode16u(byte[] p, int offset, ref UInt16 c)
        {
            UInt16 result = 0;
            result |= (UInt16)p[0 + offset];
            result |= (UInt16)(p[1 + offset] << 8);
            c = result;
            return 2;
        }

        /* encode 32 bits unsigned int (lsb) */
        public static int ikcp_encode32u(byte[] p, int offset, UInt32 l)
        {
            p[0 + offset] = (byte)(l >> 0);
            p[1 + offset] = (byte)(l >> 8);
            p[2 + offset] = (byte)(l >> 16);
            p[3 + offset] = (byte)(l >> 24);
            return 4;
        }

        /* decode 32 bits unsigned int (lsb) */
        public static int ikcp_decode32u(byte[] p, int offset, ref UInt32 c)
        {
            UInt32 result = 0;
            result |= (UInt32)p[0 + offset];
            result |= (UInt32)(p[1 + offset] << 8);
            result |= (UInt32)(p[2 + offset] << 16);
            result |= (UInt32)(p[3 + offset] << 24);
            c = result;
            return 4;
        }

        static UInt32 _imin_(UInt32 a, UInt32 b)
        {
            return a <= b ? a : b;
        }

        private static DateTime refTime = DateTime.Now;    //初始时间戳

        private static UInt32 currentMS()     //获取当前kcp运行的总时间ms
        {
            var ts = DateTime.Now.Subtract(refTime);
            return (UInt32)ts.TotalMilliseconds;
        }

        static UInt32 _imax_(UInt32 a, UInt32 b)
        {
            return a >= b ? a : b;
        }

        static UInt32 _ibound_(UInt32 lower, UInt32 middle, UInt32 upper)
        {
            return _imin_(_imax_(lower, middle), upper);
        }

        static Int32 _itimediff(UInt32 later, UInt32 earlier)
        {
            return ((Int32)(later - earlier));
        }

        // KCP Segment Definition
        internal class Segment
        {
            internal UInt32 conv = 0;    
            internal UInt32 cmd = 0;    //命令类型
            internal UInt32 frg = 0;    //分片的序号 非流式传输的话 消息会分成几个seg传输 降序排列
            internal UInt32 wnd = 0;    //可用接收窗口大小(接收窗口大小-接收队列大小)
            internal UInt32 ts = 0;    //发送时间戳 用于计算rto和rtt
            internal UInt32 sn = 0;    //分片序列号
            internal UInt32 una = 0;    //待接收的序列号 表示该号之前的序列号都收到 可以删除
            internal UInt32 rto = 0;    //重传超时时间
            internal UInt32 xmit = 0;    //记录了该报文被传输了几次
            internal UInt32 resendts = 0;    //下次超时重传该报文的时间戳
            internal UInt32 fastack = 0;    //收到ack时该分片被跳过的次数，用于快速重传
            internal UInt32 acked = 0;    //是否确认
            internal ByteBuffer data;

            private static Stack<Segment> msSegmentPool = new Stack<Segment>(32);

            public static Segment Get(int size)
            {
                lock (msSegmentPool)
                {
                    if (msSegmentPool.Count > 0)
                    {
                        var seg = msSegmentPool.Pop();
                        seg.data = ByteBuffer.Allocate(size, true);
                        return seg;
                    }
                }
                return new Segment(size);
            }

            public static void Put(Segment seg)
            {
                seg.reset();
                lock (msSegmentPool) {
                    msSegmentPool.Push(seg);
                }
            }

            private Segment(int size)
            {
                data = ByteBuffer.Allocate(size, true);
            }

            // encode a segment into buffer
            internal int encode(byte[] ptr, int offset)
            {

                var offset_ = offset;

                offset += ikcp_encode32u(ptr, offset, conv);
                offset += ikcp_encode8u(ptr, offset, (byte)cmd);
                offset += ikcp_encode8u(ptr, offset, (byte)frg);
                offset += ikcp_encode16u(ptr, offset, (UInt16)wnd);
                offset += ikcp_encode32u(ptr, offset, ts);
                offset += ikcp_encode32u(ptr, offset, sn);
                offset += ikcp_encode32u(ptr, offset, una);
                offset += ikcp_encode32u(ptr, offset, (UInt32)data.ReadableBytes);

                return offset - offset_;
            }

            internal void reset()
            {
                conv = 0;
                cmd = 0;
                frg = 0;
                wnd = 0;
                ts = 0;
                sn = 0;
                una = 0;
                rto = 0;
                xmit = 0;
                resendts = 0;
                fastack = 0;
                acked = 0;

                data.Clear();
                data.Dispose();
                data = null;
            }
        }

        internal struct ackItem
        {
            internal UInt32 sn; 
            internal UInt32 ts;
        }

        // kcp members.
        UInt32 conv;    //会话id 两端一致
        UInt32 mtu;     //最小传输单元 默认1400 最小50
        UInt32 mss;    //最大分片大小 不大于mtu
        UInt32 state;    //连接状态 0xff是断开
        UInt32 snd_una;    //第一个发送未确认的包
        UInt32 snd_nxt;    //下一个待发送包的序号
        UInt32 rcv_nxt;    //下一个待接收包的序号 接受端的收发窗口的起始序号rcv_nxt以及尾序号rcv_nxt+rcv_wnd
        UInt32 ts_recent;     //
        UInt32 ts_lastack;     //
        UInt32 ssthresh;    //拥堵窗口阈值
        Int32 rx_rttval;     //RTT变化量 代表网络抖动情况
        Int32 rx_srtt;    //smoothed round trip time平滑后RTT
        UInt32 rx_rto;     //从ACK接收延迟算出来的rto
        UInt32 rx_minrto;    //最小rto
        UInt32 snd_wnd;    //发送窗口大小
        UInt32 rcv_wnd;     //接受窗口大小
        UInt32 rmt_wnd;     //远端窗口大小
        UInt32 cwnd;     //拥堵窗口大小 动态变化
        UInt32 probe;    //探查变量 告知对方或者请求对方告知窗口大小
        UInt32 interval;    //flush的间隔
        UInt32 ts_flush;    //下一次flush的时间戳
        UInt32 nodelay;     //是否启动无延迟模式 如果无延迟 minrto会被设置为0 并且不启动阻塞控制
        UInt32 updated;    //在第一次update后置为1
        UInt32 ts_probe;    //下次探查时间戳
        UInt32 probe_wait;    //探查需要等待的时间
        UInt32 dead_link;    //最大重传次数  超过就认为连接断开
        UInt32 incr;    //可发送最大数据量

        Int32 fastresend;    //触发快速重传的重复ack个数
        Int32 nocwnd;     //是否取消拥堵控制
        Int32 stream;    //流模式

        //收发队列和缓存
        List<Segment> snd_queue = new List<Segment>(16);
        List<Segment> rcv_queue = new List<Segment>(16);
        List<Segment> snd_buf = new List<Segment>(16);
        List<Segment> rcv_buf = new List<Segment>(16);

        //待发送的ack的列表
        List<ackItem> acklist = new List<ackItem>(16);

        byte[] buffer;
        Int32 reserved; //buff中数据储备大小
        Action<byte[], int> output; // buffer, size  数据输出到外层

        // send windowd & recv window
        public UInt32 SndWnd { get { return snd_wnd; } }
        public UInt32 RcvWnd { get { return rcv_wnd; } }
        public UInt32 RmtWnd { get { return rmt_wnd; } }
        public UInt32 Mss { get { return mss; } }

        // get how many packet is waiting to be sent
        public int WaitSnd { get { return snd_buf.Count + snd_queue.Count; } }

        // internal time.
        public UInt32 CurrentMS { get { return currentMS(); } }

        // log
        Action<string> writelog = null;

        public const Int32 IKCP_LOG_OUTPUT = 1;
        public const Int32 IKCP_LOG_INPUT = 2;
        public const Int32 IKCP_LOG_SEND = 4;
        public const Int32 IKCP_LOG_RECV = 8;
        public const Int32 IKCP_LOG_IN_DATA = 16;
        public const Int32 IKCP_LOG_IN_ACK = 32;
        public const Int32 IKCP_LOG_IN_PROBE = 64;
        public const Int32 IKCP_LOG_IN_WINS = 128;
        public const Int32 IKCP_LOG_OUT_DATA = 256;
        public const Int32 IKCP_LOG_OUT_ACK = 512;
        public const Int32 IKCP_LOG_OUT_PROBE = 1024;
        public const Int32 IKCP_LOG_OUT_WINS = 2048;
        public Int32 logmask;

        // create a new kcp control object, 'conv' must equal in two endpoint
        // from the same connection.
        public KCP(UInt32 conv_, Action<byte[], int> output_)
        {
            conv = conv_;
            snd_wnd = IKCP_WND_SND;
            rcv_wnd = IKCP_WND_RCV;
            rmt_wnd = IKCP_WND_RCV;
            mtu = IKCP_MTU_DEF;
            mss = mtu - IKCP_OVERHEAD;
            rx_rto = IKCP_RTO_DEF;
            rx_minrto = IKCP_RTO_MIN;
            interval = IKCP_INTERVAL;
            ts_flush = IKCP_INTERVAL;
            ssthresh = IKCP_THRESH_INIT;
            dead_link = IKCP_DEADLINK;
            buffer = new byte[mtu];
            output = output_;
        }

        // 从rcvQ里面取能读到的消息大小
        public int PeekSize()
        {
            if (0 == rcv_queue.Count) return -1;

            var seq = rcv_queue[0];

            if (0 == seq.frg) return seq.data.ReadableBytes; //单个的seg或者流式传输 就直接取出data里面的数据

            if (rcv_queue.Count < seq.frg + 1) return -1; // 说明被分成几个seg的整个消息还没收到完

            int length = 0;

            foreach (var item in rcv_queue)
            {
                length += item.data.ReadableBytes;
                if (0 == item.frg)
                    break;
            }

            return length; // 收到完了 返回总消息长度
        }


        public int Recv(byte[] buffer)
        {
            return Recv(buffer, 0, buffer.Length);
        }

        // Receive data from kcp state machine
        //
        // Return number of bytes read.
        //
        // Return -1 when there is no readable data.
        //
        // Return -2 if len(buffer) is smaller than kcp.PeekSize().
        public int Recv(byte[] buffer, int index, int length)
        {
            var peekSize = PeekSize();
            if (peekSize < 0)
                return -1;

            if (peekSize > length)
                return -2;

            var fast_recover = false;
            if (rcv_queue.Count >= rcv_wnd)  // 收到消息的数量已经超过窗口的大小
                fast_recover = true;

            // 合并消息分段
            var count = 0;
            var n = index;
            foreach (var seg in rcv_queue)
            {
                // 分段消息拷贝到buffer.
                Buffer.BlockCopy(seg.data.RawBuffer, seg.data.ReaderIndex, buffer, n, seg.data.ReadableBytes);
                n += seg.data.ReadableBytes;

                count++;
                var fragment = seg.frg;

                if (ikcp_canlog(IKCP_LOG_RECV))
                {
                    ikcp_log($"recv sn={seg.sn.ToString()}");
                }

                Segment.Put(seg);
                if (0 == fragment) break;
            }

            if (count > 0)
            {
                rcv_queue.RemoveRange(0, count);
            }

            // 继续把buff中连续的消息塞到Q里面
            count = 0;
            foreach (var seg in rcv_buf)
            {
                if (seg.sn == rcv_nxt && rcv_queue.Count + count < rcv_wnd)
                {
                    rcv_queue.Add(seg);
                    rcv_nxt++;
                    count++;
                }
                else
                {
                    break;
                }
            }

            if (count > 0)
            {
                rcv_buf.RemoveRange(0, count);
            }


            // fast recover
            if (rcv_queue.Count < rcv_wnd && fast_recover)
            {
                // ready to send back IKCP_CMD_WINS in ikcp_flush
                // tell remote my window size
                probe |= IKCP_ASK_TELL;
            }

            return n - index;
        }

        public int Send(byte[] buffer)
        {
            return Send(buffer, 0, buffer.Length);
        }

        // user/upper level send, returns below zero for error
        public int Send(byte[] buffer, int index, int length)
        {
            if (0 == length) return -1;

            if (stream != 0) // 如果是流式传输 就把发送队列最后一个的data也塞满
            {
                var n = snd_queue.Count;
                if (n > 0)
                {
                    var seg = snd_queue[n - 1];
                    if (seg.data.ReadableBytes < mss)
                    {
                        var capacity = (int)(mss - seg.data.ReadableBytes);
                        var writen = Math.Min(capacity, length);
                        seg.data.WriteBytes(buffer, index, writen);
                        index += writen;
                        length -= writen;
                    }
                }
            }

            if (length == 0)
                return 0;

            var count = 0;
            if (length <= mss)
                count = 1;
            else
                count = (int)(((length) + mss - 1) / mss);

            if (count > 255) return -2;

            if (count == 0) count = 1;

            for (var i = 0; i < count; i++)
            {
                var size = Math.Min(length, (int)mss);

                var seg = Segment.Get(size);
                seg.data.WriteBytes(buffer, index, size);
                index += size;
                length -= size;

                // 流式传输就不写frg 否则一个消息分成几个seg传输 feg用于编辑数据顺序
                seg.frg = (stream == 0 ? (byte)(count - i - 1) : (byte)0); 
                snd_queue.Add(seg);
            }

            return 0;
        }

        // update ack.
        void update_ack(Int32 rtt)
        {
            // https://tools.ietf.org/html/rfc6298
            if (0 == rx_srtt)
            {
                rx_srtt = rtt;
                rx_rttval = rtt >> 1;
            }
            else
            {
                Int32 delta = rtt - rx_srtt;
                rx_srtt += (delta >> 3);
                if (0 > delta) delta = -delta;

                if (rtt < rx_srtt - rx_rttval)
                {
                    // if the new RTT sample is below the bottom of the range of
                    // what an RTT measurement is expected to be.
                    // give an 8x reduced weight versus its normal weighting
                    rx_rttval += ((delta - rx_rttval) >> 5);
                }
                else
                {
                    rx_rttval += ((delta - rx_rttval) >> 2);
                }
            }

            uint rto = (uint)(rx_srtt) + _imax_(interval, (uint)(rx_rttval) << 2);
            rx_rto = _ibound_(rx_minrto, rto, IKCP_RTO_MAX);
        }

        void shrink_buf()
        {
            if (snd_buf.Count > 0)
                snd_una = snd_buf[0].sn;
            else
                snd_una = snd_nxt;
        }

        void parse_ack(UInt32 sn)
        {
            // 如果这个ack包的sn小于una 说明收到过的ack已经超前了 sn大于nxt 讲道理 这个情况不应该出现 
            if (_itimediff(sn, snd_una) < 0 || _itimediff(sn, snd_nxt) >= 0) return;

            foreach (var seg in snd_buf)
            {
                if (sn == seg.sn)
                {
                    // 这里把这个seg加个标签 但是不删除 等到解析una的时候再删除 就不需要去动snd buf中的顺序
                    seg.acked = 1;
                    break;
                }
                if (_itimediff(sn, seg.sn) < 0) // 遍历到大于这个ack包的sn 后面的就不管了
                    break;
            }
        }

        void parse_fastack(UInt32 sn, UInt32 ts)
        {
            if (_itimediff(sn, snd_una) < 0 || _itimediff(sn, snd_nxt) >= 0)
                return;

            foreach (var seg in snd_buf)
            {
                if (_itimediff(sn, seg.sn) < 0)
                    break;
                // 这个ack包的sn比seg的sn大 但是时间比seg时间晚 说明这个seg的ack被跳过
                else if (sn != seg.sn && _itimediff(seg.ts, ts) <= 0)
                    seg.fastack++;
            }
        }

        // 解析已经ack的消息序号
        int parse_una(UInt32 una)
        {
            var count = 0;
            foreach (var seg in snd_buf)
            {
                if (_itimediff(una, seg.sn) > 0) {
                    count++;
                    Segment.Put(seg);    //已经确认的seg就回收
                }
                else
                    break;
            }

            if (count > 0)
                snd_buf.RemoveRange(0, count);    //确认的seg从发送队列中删除
            return count;
        }

        void ack_push(UInt32 sn, UInt32 ts)
        {
            acklist.Add(new ackItem { sn = sn, ts = ts });
        }

        bool parse_data(Segment newseg)
        {
            var sn = newseg.sn;
            // 超出窗口范围
            if (_itimediff(sn, rcv_nxt + rcv_wnd) >= 0 || _itimediff(sn, rcv_nxt) < 0)
                return true;
            
            // 接收队列的下标就是0-n
            var n = rcv_buf.Count - 1;
            var insert_idx = 0;    // 插入队列的idx
            var repeat = false; // 是否是重复的包
            for (var i = n; i >= 0; i--)
            {
                var seg = rcv_buf[i];
                if (seg.sn == sn)
                {
                    repeat = true;
                    break;
                }

                if (_itimediff(sn, seg.sn) > 0)
                {
                    insert_idx = i + 1; // 找到插入的位置
                    break;
                }
            }

            if (!repeat)
            {
                if (insert_idx == n + 1)
                    rcv_buf.Add(newseg);    // 加到最后
                else
                    rcv_buf.Insert(insert_idx, newseg);    // 插入中间
            }

            // 从buf中把连续的seg移动到queue里面 
            var count = 0;
            foreach (var seg in rcv_buf)
            {
                if (seg.sn == rcv_nxt && rcv_queue.Count + count < rcv_wnd)
                {
                    rcv_nxt++;
                    count++;
                }
                else
                {
                    break;
                }
            }

            if (count > 0)
            {
                for (var i = 0; i < count; i++)
                    rcv_queue.Add(rcv_buf[i]);
                rcv_buf.RemoveRange(0, count);
            }
            return repeat;
        }
        
        // 收到底层数据的时候调用这个地方
        // 设置了ackNoDelay的话 会马上ack
        public int Input(byte[] data, int index, int size, bool regular, bool ackNoDelay)
        {
            var s_una = snd_una;
            if (size < IKCP_OVERHEAD) return -1;

            Int32 offset = index;
            UInt32 latest = 0;
            int flag = 0;
            UInt64 inSegs = 0;
            bool windowSlides = false;

            if (ikcp_canlog(IKCP_LOG_INPUT))
            {
                ikcp_log($"[RI] {size.ToString()} bytes");
            }
            
            //循环取出各个seg
            while (true)
            {
                UInt32 ts = 0;
                UInt32 sn = 0;
                UInt32 length = 0;
                UInt32 una = 0;
                UInt32 conv_ = 0;
                UInt32 current = currentMS();

                UInt16 wnd = 0;
                byte cmd = 0;
                byte frg = 0;
                
                if (size - (offset - index) < IKCP_OVERHEAD) break;

                offset += ikcp_decode32u(data, offset, ref conv_);

                if (conv != conv_) return -1;

                offset += ikcp_decode8u(data, offset, ref cmd);
                offset += ikcp_decode8u(data, offset, ref frg);
                offset += ikcp_decode16u(data, offset, ref wnd);
                offset += ikcp_decode32u(data, offset, ref ts);
                offset += ikcp_decode32u(data, offset, ref sn);
                offset += ikcp_decode32u(data, offset, ref una);
                offset += ikcp_decode32u(data, offset, ref length);

                if (size - (offset - index) < length) return -2; //如果seg的data部分没收完 就return

                switch (cmd)
                {
                    case IKCP_CMD_PUSH:
                    case IKCP_CMD_ACK:
                    case IKCP_CMD_WASK:
                    case IKCP_CMD_WINS:
                        break;
                    default:
                        return -3;    //cmd不对 也return
                }

                // only trust window updates from regular packets. i.e: latest update
                if (regular)
                {
                    rmt_wnd = wnd;
                }

                if (parse_una(una) > 0) {
                    windowSlides = true; //有消息确认 说明发送窗口滑动了
                }
                
                shrink_buf(); //更新发送队列中第一个未确认的消息的序号

                if (IKCP_CMD_ACK == cmd)
                {
                    parse_ack(sn);    //解析ack
                    parse_fastack(sn, ts);    //通过ack包的sn和时间 解析发送队列中seg被跳过的次数
                    flag |= 1;
                    latest = ts;

                    if (ikcp_canlog(IKCP_LOG_IN_ACK))
                    {
                        ikcp_log($" input ack: sn={sn.ToString()} ts={ts.ToString()} rtt={_itimediff(current, ts).ToString()} rto={rx_rto.ToString()}");
                    }
                }
                else if (IKCP_CMD_PUSH == cmd)
                {
                    if (ikcp_canlog(IKCP_LOG_IN_DATA))
                    {
                        ikcp_log($" input psh: sn={sn.ToString()} ts={ts.ToString()}");
                    }

                    var repeat = true;
                    if (_itimediff(sn, rcv_nxt + rcv_wnd) < 0) // 收消息在窗口内
                    {
                        ack_push(sn, ts);    // 消息的ack进入list 等flush的时候发送对端
                        if (_itimediff(sn, rcv_nxt) >= 0)
                        {
                            var seg = Segment.Get((int)length);
                            seg.conv = conv_;
                            seg.cmd = (UInt32)cmd;
                            seg.frg = (UInt32)frg;
                            seg.wnd = (UInt32)wnd;
                            seg.ts = ts;
                            seg.sn = sn;
                            seg.una = una;
                            seg.data.WriteBytes(data, offset, (int)length);
                            repeat = parse_data(seg);    // 构建成seg来解析
                        }
                    }
                }
                else if (IKCP_CMD_WASK == cmd)
                {
                    // 修改prode 在flush里面会把窗口大小传给远端
                    probe |= IKCP_ASK_TELL;

                    if (ikcp_canlog(IKCP_LOG_IN_PROBE))
                    {
                        ikcp_log(" input probe");
                    }
                }
                else if (IKCP_CMD_WINS == cmd)
                {
                    // 啥都不干
                    if (ikcp_canlog(IKCP_LOG_IN_WINS))
                    {
                        ikcp_log($" input wins: {wnd.ToString()}");
                    }
                }
                else
                {
                    return -3;
                }

                inSegs++;
                offset += (int)length;
            }
            
            // cmd是ack的话 flag为1 latest是这个ack的ts
            if (flag != 0 && regular)
            {
                var current = currentMS();
                if (_itimediff(current, latest) >= 0)
                {
                    update_ack(_itimediff(current, latest)); //通过新的rtt 去计算新的rto
                }
            }

            // 拥堵控制
            if (nocwnd == 0)
            {
                // 有新的ack确认的话 snd_una会变大
                if (_itimediff(snd_una, s_una) > 0)
                {
                    if (cwnd < rmt_wnd)
                    {
                        var _mss = mss;
                        if (cwnd < ssthresh)
                        {
                            cwnd++;
                            incr += _mss;
                        }
                        else
                        {
                            if (incr < _mss)
                            {
                                incr = _mss;
                            }
                            incr += (_mss * _mss) / incr + (_mss) / 16;
                            if ((cwnd + 1) * _mss <= incr)
                            {
                                if (_mss > 0)
                                    cwnd = (incr + _mss - 1) / _mss;
                                else
                                    cwnd = incr + _mss - 1;
                            }
                        }
                        if (cwnd > rmt_wnd)
                        {
                            cwnd = rmt_wnd;
                            incr = rmt_wnd * _mss;
                        }
                    }
                }
            }

            if (windowSlides)   // 窗口滑动过 说明buff中有消息被确认了 需要发送消息
            {
                Flush(false);
            }
            else if (ackNoDelay && acklist.Count > 0) // 没有窗口欢动 就只处理ack
            {
                Flush(true);
            }

            return 0;
        }

        // 可用接收窗口的大小
        UInt16 wnd_unused()
        {
            if (rcv_queue.Count < rcv_wnd)
                return (UInt16)(rcv_wnd - rcv_queue.Count);
            return 0;
        }

        //
        int makeSpace(int space, int writeIndex)
        {
            // 如果请求空间超过mtu 直接发送到外层
            if (writeIndex + space > mtu) 
            {
                if (ikcp_canlog(IKCP_LOG_OUTPUT))
                {
                    ikcp_log($"[RO] {writeIndex.ToString()} bytes");
                }
                output(buffer, writeIndex);
                writeIndex = reserved;
            }
            return writeIndex;
        }

        
        void flushBuffer(int writeIndex)
        {
            if (writeIndex > reserved)
            {
                if (ikcp_canlog(IKCP_LOG_OUTPUT))
                {
                    ikcp_log($"[RO] {writeIndex.ToString()} bytes");
                }
                output(buffer, writeIndex);
            }
        }

        // 先把ack和一些命令的数据发了
        // 再来看堵不堵 需要调整滑动窗口等等
        // 数据从sendQ取出到sendbuffer
        public UInt32 Flush(bool ackOnly)
        {
            var seg = Segment.Get(32); //new出一个seg
            seg.conv = conv;
            seg.cmd = IKCP_CMD_ACK;    //确认命令
            seg.wnd = (UInt32)wnd_unused();    //可用窗口大小
            seg.una = rcv_nxt;

            var writeIndex = reserved;    //seg中数据可能有储备空间 就从储备空间的后续开始写

            // flush acknowledges
            for (var i = 0; i < acklist.Count; i++)
            {
                writeIndex = makeSpace(KCP.IKCP_OVERHEAD, writeIndex);    //超过最大mtu就发送到外层
                var ack = acklist[i];
                //遍历待ack的列表 如果此ack中的sn号大于待接收消息的序号(这里不是很懂 后面再看) 或者到底了
                if ( _itimediff(ack.sn, rcv_nxt) >=0 || acklist.Count - 1 == i)
                {
                    seg.sn = ack.sn;
                    seg.ts = ack.ts;
                    writeIndex += seg.encode(buffer, writeIndex); //idx往前推  

                    if (ikcp_canlog(IKCP_LOG_OUT_ACK))
                    {
                        ikcp_log($"output ack: sn={seg.sn.ToString()}");
                    }
                }
            }
            acklist.Clear();

            // flash remain ack segments
            if (ackOnly)
            {
                flushBuffer(writeIndex);    //剩余的buffer也发送出去
                Segment.Put(seg);    //seg池子回收掉
                return interval;
            }
            
            uint current = 0;
            // 远端接收窗口大小如果为0 就不发送消息 设置等待时间去探查
            if (0 == rmt_wnd)
            {
                current = currentMS();
                if (0 == probe_wait)
                {
                    probe_wait = IKCP_PROBE_INIT;
                    ts_probe = current + probe_wait;
                }
                else
                {
                    //探查等待时间不为0 并且下次探查的时间戳已经到了
                    if (_itimediff(current, ts_probe) >= 0)
                    {
                        if (probe_wait < IKCP_PROBE_INIT)
                            probe_wait = IKCP_PROBE_INIT;
                        probe_wait += probe_wait / 2;  //等待时间成为1.5倍 
                        if (probe_wait > IKCP_PROBE_LIMIT)
                            probe_wait = IKCP_PROBE_LIMIT;    //最大不超过120s
                        ts_probe = current + probe_wait;
                        probe |= IKCP_ASK_SEND;
                    }
                }
            }
            else
            {
                ts_probe = 0;
                probe_wait = 0;
            }

            // 探特么的
            if ((probe & IKCP_ASK_SEND) != 0)
            {
                seg.cmd = IKCP_CMD_WASK;
                writeIndex = makeSpace(IKCP_OVERHEAD, writeIndex);
                writeIndex += seg.encode(buffer, writeIndex);
            }
            
            // 如果是别人探自己 就把cmd设置成告知 窗口大小在encode里面
            if ((probe & IKCP_ASK_TELL) != 0)
            {
                seg.cmd = IKCP_CMD_WINS;
                writeIndex = makeSpace(IKCP_OVERHEAD, writeIndex);
                writeIndex += seg.encode(buffer, writeIndex);
            }

            probe = 0;

            // 看发送窗口和远程窗口哪个小 进行拥堵控制
            var cwnd_ = _imin_(snd_wnd, rmt_wnd);
            if (0 == nocwnd)
                cwnd_ = _imin_(cwnd, cwnd_);

            // 窗口滑动控制
            var newSegsCount = 0;
            for (var k = 0; k < snd_queue.Count; k++)
            {
                //如果下一次要发送的包序号大于未确认的包加窗口大小 也就是说等段发送的窗口已经满了 就先稳一下
                if (_itimediff(snd_nxt, snd_una + cwnd_) >= 0) 
                    break;

                //Q -> buffer
                var newseg = snd_queue[k];
                newseg.conv = conv;
                newseg.cmd = IKCP_CMD_PUSH;
                newseg.sn = snd_nxt;
                snd_buf.Add(newseg);
                snd_nxt++;
                newSegsCount++;
            }

            //Q中移除掉发送了的
            if (newSegsCount > 0)
            {
                snd_queue.RemoveRange(0, newSegsCount);
            }

            // 触发快速重传的重复ack个数
            var resent = (UInt32)fastresend;
            if (fastresend <= 0) resent = 0xffffffff;

            // check for retransmissions
            current = currentMS();
            UInt64 change = 0;     //跳过ack造成的重传次数
            UInt64 lostSegs = 0;     //超时重传次数
            UInt64 fastRetransSegs = 0;    //快速重传次数
            UInt64 earlyRetransSegs = 0;    //早重传次数
            var minrto = (Int32)interval;

            for (var k = 0; k < snd_buf.Count; k++)
            {
                var segment = snd_buf[k];
                var needsend = false;
                if (segment.acked == 1)    //如果已经ack 就不管了
                    continue;
                if (segment.xmit == 0)  // 初次传输
                {
                    needsend = true;
                    segment.rto = rx_rto;
                    segment.resendts = current + segment.rto;
                }
                else if (segment.fastack >= resent) // 这个seg在ack的时候被跳过的次数大于检查的次数 说明需要重传啦
                {
                    needsend = true;
                    segment.fastack = 0;
                    segment.rto = rx_rto;
                    segment.resendts = current + segment.rto;
                    change++;
                    fastRetransSegs++;
                }
                else if (segment.fastack > 0 && newSegsCount == 0) // 如果没有新的seg  老的seg被跳过了 就重发
                {
                    needsend = true;
                    segment.fastack = 0;
                    segment.rto = rx_rto;
                    segment.resendts = current + segment.rto;
                    change++;
                    earlyRetransSegs++;
                }
                else if (_itimediff(current, segment.resendts) >= 0) // 超过时间了 重传
                {
                    needsend = true;
                    if (nodelay == 0)    //设置的不延迟重传的话 下次时间减半 
                        segment.rto += rx_rto;
                    else
                        segment.rto += rx_rto / 2;
                    segment.fastack = 0;
                    segment.resendts = current + segment.rto;
                    lostSegs++;
                }
                
                if (needsend)
                {
                    current = CurrentMS;
                    segment.xmit++;
                    segment.ts = current;
                    segment.wnd = seg.wnd; //更新一下wnd和una
                    segment.una = seg.una;

                    var need = IKCP_OVERHEAD + segment.data.ReadableBytes;
                    writeIndex = makeSpace(need, writeIndex);
                    writeIndex += segment.encode(buffer, writeIndex); //这里只是考个头到buffer里面 下面才是考数据
                    Buffer.BlockCopy(segment.data.RawBuffer, segment.data.ReaderIndex, buffer, writeIndex, segment.data.ReadableBytes);
                    writeIndex += segment.data.ReadableBytes;

                    // 重发次数太多 就认为连接挂了
                    if (segment.xmit >= dead_link)
                    {
                        state = 0xFFFFFFFF;
                    }

                    if (ikcp_canlog(IKCP_LOG_OUT_DATA))
                    {
                        ikcp_log($"output psh: sn={segment.sn.ToString()} ts={segment.ts.ToString()} resendts={segment.resendts.ToString()} rto={segment.rto.ToString()} fastack={segment.fastack.ToString()}, xmit={segment.xmit.ToString()}");
                    }
                }

                // 得到最近的重传时间
                var _rto = _itimediff(segment.resendts, current);
                if (_rto > 0 && _rto < minrto)
                {
                    minrto = _rto;
                }
            }

            // 剩余的buffer也发出去
            flushBuffer(writeIndex);

            // 阻塞控制
            if (nocwnd == 0)
            {
                // update ssthresh
                // rate halving, https://tools.ietf.org/html/rfc6937
                if (change > 0)
                {
                    var inflght = snd_nxt - snd_una;
                    ssthresh = inflght / 2;
                    if (ssthresh < IKCP_THRESH_MIN)
                        ssthresh = IKCP_THRESH_MIN;
                    cwnd = ssthresh + resent;
                    incr = cwnd * mss;
                }

                // congestion control, https://tools.ietf.org/html/rfc5681
                if (lostSegs > 0)
                {
                    ssthresh = cwnd / 2;
                    if (ssthresh < IKCP_THRESH_MIN)
                        ssthresh = IKCP_THRESH_MIN;
                    cwnd = 1;
                    incr = mss;
                }

                if (cwnd < 1)
                {
                    cwnd = 1;
                    incr = mss;
                }
            }

            Segment.Put(seg);
            return (UInt32)minrto;
        }

        // kcp的主循环入口 10-100ms调用一次
        public void Update()
        {
            var current = currentMS();

            if (0 == updated)
            {
                updated = 1;
                ts_flush = current;    
            }

            var slap = _itimediff(current, ts_flush); //计算与下一次flush的距离时间

            if (slap >= 10000 || slap < -10000) //超过一秒就当场flush
            {
                ts_flush = current;
                slap = 0;
            }

            if (slap >= 0)    //已经到达下次flush的时间了
            {
                ts_flush += interval;
                if (_itimediff(current, ts_flush) >= 0)
                    ts_flush = current + interval;
                Flush(false);
            }
        }

        // Determine when should you invoke ikcp_update:
        // returns when you should invoke ikcp_update in millisec, if there
        // is no ikcp_input/_send calling. you can call ikcp_update in that
        // time, instead of call update repeatly.
        // Important to reduce unnacessary ikcp_update invoking. use it to
        // schedule ikcp_update (eg. implementing an epoll-like mechanism,
        // or optimize ikcp_update when handling massive kcp connections)
        public UInt32 Check()
        {
            var current = currentMS();

            var ts_flush_ = ts_flush;
            var tm_flush_ = 0x7fffffff;
            var tm_packet = 0x7fffffff;
            var minimal = 0;

            if (updated == 0)
                return current;

            if (_itimediff(current, ts_flush_) >= 10000 || _itimediff(current, ts_flush_) < -10000)
                ts_flush_ = current;

            if (_itimediff(current, ts_flush_) >= 0)
                return current;

            tm_flush_ = (int)_itimediff(ts_flush_, current);

            foreach (var seg in snd_buf)
            {
                var diff = _itimediff(seg.resendts, current);
                if (diff <= 0)
                    return current;
                if (diff < tm_packet)
                    tm_packet = (int)diff;
            }

            minimal = (int)tm_packet;
            if (tm_packet >= tm_flush_)
                minimal = (int)tm_flush_;
            if (minimal >= interval)
                minimal = (int)interval;

            return current + (UInt32)minimal;
        }

        // change MTU size, default is 1400
        public int SetMtu(Int32 mtu_)
        {
            if (mtu_ < 50 || mtu_ < (Int32)IKCP_OVERHEAD)
                return -1;
            if (reserved >= (int)(mtu - IKCP_OVERHEAD) || reserved < 0)
                return -1;

            var buffer_ = new byte[mtu_];
            if (null == buffer_)
                return -2;

            mtu = (UInt32)mtu_;
            mss = mtu - IKCP_OVERHEAD - (UInt32)reserved;
            buffer = buffer_;
            return 0;
        }

        // fastest: ikcp_nodelay(kcp, 1, 20, 2, 1)
        // nodelay: 0:disable(default), 1:enable
        // interval: internal update timer interval in millisec, default is 100ms
        // resend: 0:disable fast resend(default), 1:enable fast resend
        // nc: 0:normal congestion control(default), 1:disable congestion control
        public int NoDelay(int nodelay_, int interval_, int resend_, int nc_)
        {

            if (nodelay_ >= 0)
            {
                nodelay = (UInt32)nodelay_;
                if (nodelay_ != 0)
                    rx_minrto = IKCP_RTO_NDL;
                else
                    rx_minrto = IKCP_RTO_MIN;
            }

            if (interval_ >= 0)
            {
                if (interval_ > 5000)
                    interval_ = 5000;
                else if (interval_ < 10)
                    interval_ = 10;
                interval = (UInt32)interval_;
            }

            if (resend_ >= 0)
                fastresend = resend_;

            if (nc_ >= 0)
                nocwnd = nc_;

            return 0;
        }

        // set maximum window size: sndwnd=32, rcvwnd=32 by default
        public int WndSize(int sndwnd, int rcvwnd)
        {
            if (sndwnd > 0)
                snd_wnd = (UInt32)sndwnd;

            if (rcvwnd > 0)
                rcv_wnd = (UInt32)rcvwnd;
            return 0;
        }

        //设置储备大小
        public bool ReserveBytes(int reservedSize)
        {
            if (reservedSize >= (mtu - IKCP_OVERHEAD) || reservedSize < 0)
                return false;

            reserved = reservedSize;
            mss = mtu - IKCP_OVERHEAD - (uint)(reservedSize);
            return true;
        }

        public void SetStreamMode(bool enabled)
        {
            stream = enabled ? 1 : 0;
        }

        bool ikcp_canlog(int mask)
        {
            if ((mask & logmask) == 0 || writelog == null) return false;
            return true;
        }

        public void SetLogger(Action<string> logger)
        {
            writelog = logger;
        }

        public void SetLogMask(int mask)
        {
            logmask = mask;
        }

        void ikcp_log(string logStr)
        {
            writelog?.Invoke(logStr);
        }

        public UInt32 State()
        {
            return state;
        }
    }
}
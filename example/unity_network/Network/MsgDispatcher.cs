using System;
using System.Collections.Generic;

namespace GameLogic
{
    public class MsgDispatcher : Singleton<MsgDispatcher>
    {
        private readonly Dictionary<Type, Action<object>> _receiveActionDic = new();

        /// <summary>
        /// 注册消息接收者
        /// </summary>
        public void RegisterMsgReceiver<T>(Action<T> action)
        {
            var t = typeof(T);
            if (_receiveActionDic.ContainsKey(t))
            {
                Log.Warning($"RegisterMsgReceiver err: repeat type: {t.Name}");
                return;
            }

            // 包一层转换，保证存储为 Action<object>
            _receiveActionDic.Add(t, o => action((T)o));
        }

        /// <summary>
        /// 分发消息（根据运行时类型）
        /// </summary>
        public void DispatchMsg(object msg)
        {
            if (msg == null)
            {
                Log.Warning("DispatchMsg err: msg is null");
                return;
            }

            var t = msg.GetType(); // 获取运行时真实类型

            if (_receiveActionDic.TryGetValue(t, out var recv))
            {
                recv.Invoke(msg);
                return;
            }

            Log.Warning($"DispatchMsg err: not register type: {t.Name}");
        }
    }
}
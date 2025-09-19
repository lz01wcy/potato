using System;
using System.Collections.Generic;
using UnityEngine;

namespace GameLogic
{
    public class MsgMeta
    {
        public static Dictionary<Type, int> type2Id = new Dictionary<Type, int>();
        public static Dictionary<int, Type> id2Type = new Dictionary<int, Type>();

        public static void RegisterMsgMeta(int id, Type type)
        {
            if (type2Id.ContainsKey(type))
            {
                Debug.LogError("RegisterMsgMeta err: type is repeat " + type.Name);
                return;
            }
            type2Id[type] = id;
            
            if (id2Type.ContainsKey(id))
            {
                Debug.LogError("RegisterRecvMsgMeta err: id is repeat " + id);
                return;
            }
            id2Type[id] = type;
        }
        
        public static Type GetTypeById(int id)
        {
            Type t;
            id2Type.TryGetValue(id, out t);
            if (t == null)
            {
                Debug.LogError("can not get msg type by id");
            }
            return t;
        }

        public static int GetIdByType(Type t)
        {
            int id;
            type2Id.TryGetValue(t, out id);
            if (id == 0)
            {
                Debug.LogError("can not get msg id by type");
            }
            return id;
        }
    }
}


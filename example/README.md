* server 用于客户端连接的服务器
* calculator 用于计算的服务 server节点通过rpc调用calculator节点进行计算
* client 客户端
* nicepb 使用protobuf所需的消息文件
* pairpb 作为消息对注册的protobuf
* config 配置文件读取示例

---

#### unity_network unity网络示例 
* 使用了[UniTask](https://github.com/Cysharp/UniTask)以及[UnityWebSocket](https://github.com/psygames/UnityWebSocket)适配微信小游戏的单线程限制
* 把Network文件夹拖到自己项目中 修改成自己的命名空间和Log的编译错误 基本上算是“开箱即用”
* pb文件夹中有protobuf文件的示例以及客户端的消息与id自动映射的工具以及脚本 可根据需求自己修改即可
* Useage.cs中是简单的使用示例
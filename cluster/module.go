package cluster

import (
	"fmt"
	"github.com/asynkron/protoactor-go/cluster"
	"github.com/asynkron/protoactor-go/cluster/clusterproviders/consul"
	"github.com/asynkron/protoactor-go/cluster/identitylookup/disthash"
	"github.com/asynkron/protoactor-go/remote"
	"github.com/hashicorp/consul/api"
	"github.com/murang/potato/app"
	"github.com/murang/potato/log"
	"github.com/murang/potato/util"
)

type Module struct {
	cls *cluster.Cluster // 集群

	ClusterName string        // 集群名称
	Consul      string        // 服务发现注册地址
	ServiceKind *cluster.Kind // 使用proto actor grain生成的服务类型
}

func (m *Module) FPS() uint {
	return 0
}

func (m *Module) OnStart() {
	provider, _ := consul.NewWithConfig(&api.Config{
		Address: m.Consul,
	})
	lookup := disthash.New()
	lanIp, err := util.GetLocalEthernetIP()
	if err != nil {
		log.Sugar.Errorf("GetLocalEthernetIP err: %s", err)
		return
	}
	availablePort, err := util.GetAvailablePort(40000, 50000)
	if err != nil {
		log.Sugar.Errorf("GetAvailablePort err: %s", err)
		return
	}
	config := remote.Configure(lanIp, availablePort, remote.WithAdvertisedHost(fmt.Sprintf("%s:%d", lanIp, availablePort)))
	clusterConfig := cluster.Configure(m.ClusterName, provider, lookup, config, cluster.WithKinds(m.ServiceKind))

	m.cls = cluster.New(app.Instance().GetActorSystem(), clusterConfig)
	m.cls.StartMember()
}

func (m *Module) OnUpdate() {
}

func (m *Module) OnDestroy() {
	m.cls.Shutdown(true)
}

func (m *Module) OnMsg(msg interface{}) {
}

func (m *Module) OnRequest(msg interface{}) interface{} {
	return nil
}

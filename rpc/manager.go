package rpc

import (
	"fmt"
	"github.com/asynkron/protoactor-go/actor"
	"github.com/asynkron/protoactor-go/cluster"
	"github.com/asynkron/protoactor-go/cluster/clusterproviders/consul"
	"github.com/asynkron/protoactor-go/cluster/identitylookup/disthash"
	"github.com/asynkron/protoactor-go/remote"
	"github.com/hashicorp/consul/api"
	"github.com/murang/potato/log"
	"github.com/murang/potato/util"
)

type Config struct {
	ClusterName string          // 集群名称
	Consul      string          // 服务发现注册地址
	ServiceKind []*cluster.Kind // 使用proto actor grain生成的服务类型
}

type Manager struct {
	cluster      *cluster.Cluster
	clusterName  string          // 集群名称
	consul       string          // 服务发现注册地址
	serviceKinds []*cluster.Kind // 使用proto actor grain生成的服务类型
}

func NewManagerWithConfig(config *Config) *Manager {
	return &Manager{
		clusterName:  config.ClusterName,
		consul:       config.Consul,
		serviceKinds: config.ServiceKind,
	}
}

func (m *Manager) GetCluster() *cluster.Cluster {
	return m.cluster
}

func (m *Manager) Start(actorSystem *actor.ActorSystem) (cls *cluster.Cluster) {
	provider, _ := consul.NewWithConfig(&api.Config{
		Address: m.consul,
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
	if m.serviceKinds == nil {
		clusterConfig := cluster.Configure(m.clusterName, provider, lookup, config)
		m.cluster = cluster.New(actorSystem, clusterConfig)
	} else {
		clusterConfig := cluster.Configure(m.clusterName, provider, lookup, config, cluster.WithKinds(m.serviceKinds...))
		m.cluster = cluster.New(actorSystem, clusterConfig)
	}

	m.cluster.StartMember()
	return m.cluster
}

func (m *Manager) OnDestroy() {
	m.cluster.Shutdown(true)
}

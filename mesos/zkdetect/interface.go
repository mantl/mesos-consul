package zkdetect

import (
	mesos "github.com/mesos/mesos-go/mesosproto"
)

type ClusterInfo struct {
	Leader		*mesos.MasterInfo
	Masters		*[]*mesos.MasterInfo
}

type ClusterChanged interface {
	OnClusterChanged(*ClusterInfo)
}

type OnClusterChanged func(*ClusterInfo)

func (f OnClusterChanged) OnClusterChanged(ci *ClusterInfo) {
	f(ci)
}

type Cluster interface {
	Detect(ClusterChanged) error
	Done() <-chan struct{}
	Cancel()
}

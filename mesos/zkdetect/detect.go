/**
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package zkdetect

import (
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gogo/protobuf/proto"
	log "github.com/golang/glog"
	mesos "github.com/mesos/mesos-go/mesosproto"
)

const (
	// prefix for nodes listed at the ZK URL path
	nodePrefix = "info_"
)

// reasonable default for a noop change listener
var ignoreChanged = OnClusterChanged(func(*ClusterInfo) {})

// Detector uses ZooKeeper to detect new leading master.
type ClusterDetector struct {
	client          *Client
	leaderNode      string
	bootstrap       sync.Once // for one-time zk client initiation
	ignoreInstalled int32     // only install, at most, one ignoreChanged listener; see MasterDetector.Detect
}

// Internal constructor function
func NewClusterDetector(zkurls string) (*ClusterDetector, error) {
	zkHosts, zkPath, err := parseZk(zkurls)
	if err != nil {
		log.Fatalln("Failed to parse url", err)
		return nil, err
	}

	client, err := newClient(zkHosts, zkPath)
	if err != nil {
		return nil, err
	}

	detector := &ClusterDetector{
		client: client,
	}

	log.V(2).Infoln("Created new detector, watching ", zkHosts, zkPath)
	return detector, nil
}

func parseZk(zkurls string) ([]string, string, error) {
	u, err := url.Parse(zkurls)
	if err != nil {
		log.V(1).Infof("failed to parse url: %v", err)
		return nil, "", err
	}
	if u.Scheme != "zk" {
		return nil, "", fmt.Errorf("invalid url scheme for zk url: '%v'", u.Scheme)
	}
	return strings.Split(u.Host, ","), u.Path, nil
}

// returns a chan that, when closed, indicates termination of the detector
func (md *ClusterDetector) Done() <-chan struct{} {
	return md.client.stopped()
}

func (md *ClusterDetector) Cancel() {
	md.client.stop()
}

//TODO(jdef) execute async because we don't want to stall our client's event loop? if so
//then we also probably want serial event delivery (aka. delivery via a chan) but then we
//have to deal with chan buffer sizes .. ugh. This is probably the least painful for now.
func (md *ClusterDetector) childrenChanged(zkc *Client, path string, obs ClusterChanged) {
	log.V(2).Infof("fetching children at path '%v'", path)
	list, err := zkc.list(path)
	if err != nil {
		log.Warning(err)
		return
	}

	// topNode is the leader node
	topNode := selectTopNode(list)

if topNode == md.leaderNode {
	fmt.Println("Ignoring children changed event")
} else {
	fmt.Println("Changing leader node")
}

	md.leaderNode = topNode

	var clusterInfo = new(ClusterInfo)

	clusterInfo.Masters = new([]*mesos.MasterInfo)

	for _, v := range list {
		if (!strings.HasPrefix(v, nodePrefix)) {
			continue
		}

		seqStr := strings.TrimPrefix(v, nodePrefix)
		_, err := strconv.ParseUint(seqStr, 10, 64)
		if err != nil {
			log.Warning("unexpected zk node format '%s': %v", seqStr, err)
			continue
		}

		data, err := zkc.data(fmt.Sprintf("%s/%s", path, v))
		if err != nil {
			log.Errorf("unable to retrieve master data: %v", err.Error())
			return
		}

		masterInfo := new(mesos.MasterInfo)
		err = proto.Unmarshal(data, masterInfo)
		if err != nil {
			log.Errorf("unable to unmarshall MasterInfo data from zookeeper: %v", err)
		}

		if v == md.leaderNode {
			clusterInfo.Leader = masterInfo
		} else {
			*clusterInfo.Masters = append(*clusterInfo.Masters, masterInfo)
		}
	}

	log.V(2).Infof("detected cluster info: %+v",clusterInfo)
	obs.OnClusterChanged(clusterInfo)
}

// the first call to Detect will kickstart a connection to zookeeper. a nil change listener may
// be spec'd, result of which is a detector that will still listen for master changes and record
// leaderhip changes internally but no listener would be notified. Detect may be called more than
// once, and each time the spec'd listener will be added to the list of those receiving notifications.
func (md *ClusterDetector) Detect(f ClusterChanged) (err error) {
	// kickstart zk client connectivity
	md.bootstrap.Do(func() { go md.client.connect() })

	if f == nil {
		// only ever install, at most, one ignoreChanged listener. multiple instances of it
		// just consume resources and generate misleading log messages.
		if !atomic.CompareAndSwapInt32(&md.ignoreInstalled, 0, 1) {
			return
		}
		f = ignoreChanged
	}

	go md.detect(f)
	return nil
}

func (md *ClusterDetector) detect(f ClusterChanged) {

	minCyclePeriod := 1 * time.Second
detectLoop:
	for {
		started := time.Now()
		select {
		case <-md.Done():
			return
		case <-md.client.connections():
			// we let the golang runtime manage our listener list for us, in form of goroutines that
			// callback to the master change notification listen func's
			if watchEnded, err := md.client.watchChildren(currentPath, ChildWatcher(func(zkc *Client, path string) {
				md.childrenChanged(zkc, path, f)
			})); err == nil {
				log.V(2).Infoln("detector listener installed")
				select {
				case <-watchEnded:
					if md.leaderNode != "" {
						log.V(1).Infof("child watch ended, signaling master lost")
						md.leaderNode = ""
						f.OnClusterChanged(nil)
					}
				case <-md.client.stopped():
					return
				}
			} else {
				log.V(1).Infof("child watch ended with error: %v", err)
				continue detectLoop
			}
		}
		// rate-limit master changes
		if elapsed := time.Now().Sub(started); elapsed > 0 {
			log.V(2).Infoln("resting before next detection cycle")
			select {
			case <-md.Done():
				return
			case <-time.After(minCyclePeriod - elapsed): // noop
			}
		}
	}
}

func selectTopNode(list []string) (node string) {
	var leaderSeq uint64 = math.MaxUint64

	for _, v := range list {
		if !strings.HasPrefix(v, nodePrefix) {
			continue // only care about participants
		}
		seqStr := strings.TrimPrefix(v, nodePrefix)
		seq, err := strconv.ParseUint(seqStr, 10, 64)
		if err != nil {
			log.Warningf("unexpected zk node format '%s': %v", seqStr, err)
			continue
		}
		if seq < leaderSeq {
			leaderSeq = seq
			node = v
		}
	}

	if node == "" {
		log.V(3).Infoln("No top node found.")
	} else {
		log.V(3).Infof("Top node selected: '%s'", node)
	}
	return node
}

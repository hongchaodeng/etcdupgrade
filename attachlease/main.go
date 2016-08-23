/*
Copyright 2016 The Kubernetes Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/golang/glog"
	"golang.org/x/net/context"
)

func main() {
	var ttlDir string
	var etcdHost string
	var leaseDuration time.Duration

	flag.StringVar(&ttlDir, "ttldir", "", "ttl keys directory")
	flag.StringVar(&etcdHost, "etcd-addr", "", "Etcd address")
	flag.DurationVar(&leaseDuration, "lease-duration", time.Hour, "Lease duration")
	flag.Parse()

	if ttlDir == "" {
		glog.Fatalf("--ttldir flag is required")
	}
	if etcdHost == "" {
		glog.Fatalf("--etcd-addr flag is required")
	}

	client, err := clientv3.New(clientv3.Config{Endpoints: []string{etcdHost}})
	if err != nil {
		glog.Fatalf("Error while creating etcd client: %v", err)
	}
	lease, err := client.Lease.Grant(context.TODO(), int64(leaseDuration/time.Second))
	if err != nil {
		glog.Fatalf("Error while creating lease: %v", err)
	}

	if strings.HasSuffix(ttlDir, "/") {
		ttlDir = ttlDir + "/"
	}
	// TODO: pagination
	getResp, err := client.KV.Get(context.TODO(), ttlDir, clientv3.WithPrefix())
	if err != nil {
		panic(err)
	}
	for _, kv := range getResp.Kvs {
		_, err := client.KV.Put(context.TODO(), string(kv.Key), string(kv.Value), clientv3.WithLease(lease.ID))
		if err != nil {
			panic(err)
		}
	}
}

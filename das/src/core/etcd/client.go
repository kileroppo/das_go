package etcd

import (
	"strings"
	"sync"
	"time"

	"github.com/etcd-io/etcd/clientv3"
	"google.golang.org/grpc"

	"das/core/log"
)

var etcdCli *clientv3.Client
var etcdOnce sync.Once

func Init() {
	etcdOnce.Do(func() {
		etcdCli = newEtcdClient()
	})
}

func CloseEtcdCli() {
	if etcdCli != nil {
	    etcdCli.Close()
	}
}

func GetEtcdClient() *clientv3.Client {
	return etcdCli
}

func newEtcdClient() *clientv3.Client {
	rawHost,err := log.Conf.GetString("etcd", "url")
	if err != nil {
		panic(err)
	}
	hosts := strings.Split(rawHost, ";")

	cfg := clientv3.Config{
		Endpoints:           hosts,
		AutoSyncInterval:    0,
		DialTimeout:         time.Second * 3,
		Username:            "",
		Password:            "",
		RejectOldCluster:    false,
		DialOptions:         []grpc.DialOption{grpc.WithBlock()},
	}

	cli,err := clientv3.New(cfg)
	if err != nil {
		log.Errorf("newEtcdClient > clientv3.New > %s", err)
		panic(err)
	}

	return cli
}
package etcd

import (
	"context"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdClient struct {
	client *clientv3.Client
}

func NewEtcdClient(endpoints []string) (*EtcdClient, error) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	return &EtcdClient{client: etcdClient}, nil
}

func (etcdClient *EtcdClient) ResetEtcdData() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := etcdClient.client.Delete(ctx, "/", clientv3.WithPrefix())
	if err != nil {
		return err
	}
	return nil
}

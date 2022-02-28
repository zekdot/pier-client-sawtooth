package main

import "net/rpc"

type RpcClient struct {
	client *rpc.Client
}

type ReqArgs struct {
	FuncName string
	Args []string
}

func (rpcClient *RpcClient) GetData(key string) string {

}

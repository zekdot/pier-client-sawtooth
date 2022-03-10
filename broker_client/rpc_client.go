package main

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc"
	pb "nju.edu.cn/zekdot/cli/envelope"
	"strconv"
	"time"
)

type RpcClient struct {
	client *pb.PostServiceClient
	ctx *context.Context
}

func NewRpcClient(address string) (*RpcClient, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	//defer conn.Close()
	client := pb.NewPostServiceClient(conn)
	//client.SendEnvelope()
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	//r, err := client.SendEnvelope(ctx, &pb.EnvelopeRequest{Func: "getValue", Params: []string{"outter-meta"}})
	//log.Printf(r.Result)
	//defer cancel()

	return &RpcClient{
		client: &client,
		ctx: &ctx,
	}, nil
}

func (rpcClient *RpcClient) GetData(key string) (string, error) {
	client := *rpcClient.client
	r, err := client.SendEnvelope(*rpcClient.ctx, &pb.EnvelopeRequest{Func: "getValue", Params: []string{key}})
	if err != nil {
		return "", err
	}
	return r.Result, nil
}

func (rpcClient *RpcClient) SetData(key string, value string) error {
	client := *rpcClient.client
	_, err := client.SendEnvelope(*rpcClient.ctx, &pb.EnvelopeRequest{Func: "setValue", Params: []string{key, value}})
	if err != nil {
		return err
	}
	return nil
}

func (rpcClient *RpcClient) Init() error {
	err := rpcClient.SetData("inner-meta", "{}")
	if err != nil {
		return err
	}
	err = rpcClient.SetData("outter-meta", "{}")
	if err != nil {
		return err
	}
	err = rpcClient.SetData("callback-meta", "{}")
	return nil
}

type Event struct {
	Index         uint64 `json:"index"`
	DstChainID    string `json:"dst_chain_id"`
	SrcContractID string `json:"src_contract_id"`
	DstContractID string `json:"dst_contract_id"`
	Func          string `json:"func"`
	Args          string `json:"args"`
	Callback      string `json:"callback"`
}

func (rpcClient *RpcClient) InterchainGet(toId string, contractId string, key string) error {
	cid := "mychannel&data_swapper"
	//destChainId := toId
	outMetaStr, err := rpcClient.GetData("outter-meta")
	if err != nil {
		return err
	}
	outMeta := make(map[string]uint64)
	err = json.Unmarshal([]byte(outMetaStr), &outMeta)
	if err != nil {
		return err
	}
	if _, ok := outMeta[toId]; !ok {
		outMeta[toId] = 0
	}
	tx := &Event{
		Index:         outMeta[toId] + 1,
		DstChainID:    toId,
		SrcContractID: cid,
		DstContractID: contractId,
		Func:          "interchainGet",
		Args:          key,
		Callback:      "interchainSet",
	}
	outMeta[toId] ++
	outMetaBytes, err := json.Marshal(outMeta)
	if err != nil {
		return err
	}
	err = rpcClient.SetData("outter-meta", string(outMetaBytes))
	if err != nil {
		return err
	}
	txValue, err := json.Marshal(tx)
	if err != nil {
		return err
	}
	outKey := outMsgKey(tx.DstChainID, strconv.FormatUint(tx.Index, 10))
	err = rpcClient.SetData(outKey, string(txValue))
	if err != nil {
		return err
	}
	return nil
}

func outMsgKey(to string, idx string) string {
	return fmt.Sprintf("out-msg-%s-%s", to, idx)
}
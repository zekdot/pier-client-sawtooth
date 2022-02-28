package main

import (
	"encoding/json"
	"fmt"
	"net/rpc"
	"strconv"
	"strings"
)

type RpcClient struct {
	client *rpc.Client
}

type ReqArgs struct {
	FuncName string
	Args []string
}

func NewRpcClient(address string) (*RpcClient, error) {
	rpcClient, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		return nil, err
	}
	return &RpcClient{
		client: rpcClient,
	}, nil
}

func (rpcClient *RpcClient) polling(m map[string]uint64) ([]*Event, error) {
	outMeta, err := rpcClient.GetOuterMeta()
	if err != nil {
		return nil, err
	}
	events := make([]*Event, 0)
	for addr, idx := range outMeta {
		startPos, ok := m[addr]
		if !ok {
			startPos = 0
		}
		for i := startPos + 1; i <= idx; i++ {
			event, _ := rpcClient.GetOutMessage(addr, i)

			events = append(events, event)
		}
	}
	return events, nil
}

func (rpcClient *RpcClient) getMessage(key string) ([][]byte, error) {
	var reply string
	reqArgs := ReqArgs{
		"get",
		[]string{key},
	}
	err := rpcClient.client.Call("Service.GetValue", reqArgs, &reply)
	if err != nil {
		return nil, err
	}
	results := strings.Split(reply, ",")
	return toChaincodeArgs(results...), nil
}

func (rpcClient *RpcClient) GetInMessage(sourceChainID string, sequenceNum uint64)([][]byte, error) {
	key := inMsgKey(sourceChainID, strconv.FormatUint(sequenceNum, 10))
	return rpcClient.getMessage(key)
}

func (rpcClient *RpcClient) GetOutMessage(sourceChainID string, sequenceNum uint64)(*Event, error) {
	key := inMsgKey(sourceChainID, strconv.FormatUint(sequenceNum, 10))
	var reply string
	reqArgs := ReqArgs{
		"get",
		[]string{key},
	}
	err := rpcClient.client.Call("Service.GetValue", reqArgs, &reply)
	if err != nil {
		return nil, err
	}
	ret := &Event{}
	if err := json.Unmarshal([]byte(reply), ret); err != nil {
		return nil, err
	}
	return ret, nil
}

// ToChaincodeArgs converts string args to []byte args
func toChaincodeArgs(args ...string) [][]byte {
	bargs := make([][]byte, len(args))
	for i, arg := range args {
		bargs[i] = []byte(arg)
	}
	return bargs
}

func (rpcClient *RpcClient) getMeta(key string) (map[string]uint64, error) {
	var reply string
	reqArgs := ReqArgs{
		"get",
		[]string{key},
	}
	err := rpcClient.client.Call("Service.GetValue", reqArgs, &reply)
	if err != nil {
		return nil, err
	}
	outMeta := make(map[string]uint64)
	err = json.Unmarshal([]byte(reply), &outMeta)
	if err != nil {
		return nil, err
	}
	return outMeta, nil
}

func (rpcClient *RpcClient) GetInnerMeta() (map[string]uint64, error) {
	return rpcClient.getMeta("inner-meta")
}

func (rpcClient *RpcClient) GetOuterMeta() (map[string]uint64, error) {
	return rpcClient.getMeta("outter-meta")
}

func (rpcClient *RpcClient) GetCallbackMeta() (map[string]uint64, error) {
	return rpcClient.getMeta("callback-meta")
}

func (rpcClient *RpcClient) Init() error {
	var reply string
	reqArgs := ReqArgs{
		"init",
		[]string{"inner-meta", "{}"},
	}
	err := rpcClient.client.Call("Service.SetValue", reqArgs, &reply)
	if err != nil {
		return err
	}
	reqArgs = ReqArgs{
		"init",
		[]string{"outter-meta", "{}"},
	}
	err = rpcClient.client.Call("Service.SetValue", reqArgs, &reply)
	if err != nil {
		return err
	}
	reqArgs = ReqArgs{
		"init",
		[]string{"callback-meta", "{}"},
	}
	err = rpcClient.client.Call("Service.SetValue", reqArgs, &reply)
	if err != nil {
		return err
	}
	return nil
}

func (rpcClient *RpcClient) GetData(key string) (string, error) {
	var reply string
	reqArgs := ReqArgs{
		"get",
		[]string{key},
	}
	err := rpcClient.client.Call("Service.GetValue", reqArgs, &reply)
	if err != nil {
		return "", err
	}
	return reply, nil
}

func (rpcClient *RpcClient) SetData(key string, value string) error {
	var reply string
	reqArgs := ReqArgs{
		"set",
		[]string{key, value},
	}
	err := rpcClient.client.Call("Service.SetValue", reqArgs, &reply)
	if err != nil {
		return err
	}
	return nil
}

func (rpcClient *RpcClient) InterchainGet(toId string, contractId string, key string) error {
	cid := "mychannel&data_swapper"
	//destChainId := toId
	var reply string
	reqArgs := ReqArgs{
		"get",
		[]string{"outter-meta"},
	}
	err := rpcClient.client.Call("Service.GetValue", reqArgs, &reply)
	if err != nil {
		return err
	}
	outMeta := make(map[string]uint64)
	err = json.Unmarshal([]byte(reply), &outMeta)
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
	reqArgs = ReqArgs{
		"set",
		[]string{"outter-meta", string(outMetaBytes)},
	}
	err = rpcClient.client.Call("Service.SetValue", reqArgs, &reply)
	if err != nil {
		return err
	}
	txValue, err := json.Marshal(tx)
	if err != nil {
		return err
	}
	outKey := outMsgKey(tx.DstChainID, strconv.FormatUint(tx.Index, 10))
	reqArgs = ReqArgs{
		"set",
		[]string{outKey, string(txValue)},
	}
	err = rpcClient.client.Call("Service.SetValue", reqArgs, &reply);
	if err != nil {
		return err
	}
	return nil
}

func outMsgKey(to string, idx string) string {
	return fmt.Sprintf("out-msg-%s-%s", to, idx)
}

func inMsgKey(to string, idx string) string {
	return fmt.Sprintf("in-msg-%s-%s", to, idx)
}
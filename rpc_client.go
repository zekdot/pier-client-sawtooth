package main

import (
	"encoding/json"
	"fmt"
	"net/rpc"
	"strconv"
	"strings"
)

const (
	delimiter = "&"
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

func (rpcClient *RpcClient) InterchainSet(args[] string) error {
	if len(args) < 5 {
		return fmt.Errorf("incorrect number of arguments, expecting 5")
	}
	sourceChainID := args[0]
	sequenceNum := args[1]
	targetCID := args[2]
	key := args[3]
	data := args[4]

	if err := rpcClient.checkIndex(sourceChainID, sequenceNum, "callback-meta"); err != nil {
		return err
	}

	idx, err := strconv.ParseUint(sequenceNum, 10, 64)
	if err != nil {
		return err
	}

	if err := rpcClient.markCallbackCounter(sourceChainID, idx); err != nil {
		return err
	}

	splitedCID := strings.Split(targetCID, delimiter)
	if len(splitedCID) != 2 {
		return fmt.Errorf("Target chaincode id %s is not valid", targetCID)
	}

	var reply string
	reqArgs := ReqArgs{
		"set",
		[]string{key, data},
	}
	err = rpcClient.client.Call("Service.SetValue", reqArgs, &reply)
	if err != nil {
		return err
	}

	return nil
}

func (rpcClient *RpcClient) InterchainGet(args[] string) (string, error) {

	if len(args) < 4 {
		return "", fmt.Errorf("incorrect number of arguments, expecting 4")
	}
	sourceChainID := args[0]
	sequenceNum := args[1]
	targetCID := args[2]
	key := args[3]

	if err := rpcClient.checkIndex(sourceChainID, sequenceNum, "inner-meta"); err != nil {
		return "", err
	}

	if err := rpcClient.markInCounter(sourceChainID); err != nil {
		return "", err
	}

	splitedCID := strings.Split(targetCID, delimiter)
	if len(splitedCID) != 2 {
		return "", fmt.Errorf("Target chaincode id %s is not valid", targetCID)
	}

	// args[0]: key
	value, err := rpcClient.GetData(key)
	if err != nil {
		return "", err
	}

	inKey := inMsgKey(sourceChainID, sequenceNum)
	if err := rpcClient.SetData(inKey, value); err != nil {
		return "", err
	}

	return value, nil
}

func (rpcClient *RpcClient) markInCounter(from string) error {
	inMeta, err := rpcClient.GetInnerMeta()
	if err != nil {
		return err
	}

	inMeta[from]++
	metaStr, err := json.Marshal(inMeta)
	if err != nil {
		return err
	}
	var reply string
	reqArgs := ReqArgs{
		"set",
		[]string{"inner-meta", string(metaStr)},
	}
	err = rpcClient.client.Call("Service.SetValue", reqArgs, &reply)
	return err
}

func (rpcClient *RpcClient) markCallbackCounter(from string, index uint64) error {
	meta, err := rpcClient.GetCallbackMeta()
	if err != nil {
		return err
	}

	meta[from] = index
	metaStr, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	var reply string
	reqArgs := ReqArgs{
		"set",
		[]string{"callback-meta", string(metaStr)},
	}
	err = rpcClient.client.Call("Service.SetValue", reqArgs, &reply)
	return err
}

func (rpcClient *RpcClient) checkIndex(addr string, index string, metaName string) error {
	idx, err := strconv.ParseUint(index, 10, 64)
	if err != nil {
		return err
	}
	meta, err := rpcClient.getMeta(metaName)  //broker.getMap(state, metaName)
	if err != nil {
		return err
	}
	if idx != meta[addr] + 1 {
		return fmt.Errorf("incorrect index, expect %d", meta[addr]+1)
	}
	return nil
}

func (rpcClient *RpcClient) Polling(m map[string]uint64) ([]*Event, error) {
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
	key := outMsgKey(sourceChainID, strconv.FormatUint(sequenceNum, 10))
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

func outMsgKey(to string, idx string) string {
	return fmt.Sprintf("out-msg-%s-%s", to, idx)
}

func inMsgKey(to string, idx string) string {
	return fmt.Sprintf("in-msg-%s-%s", to, idx)
}
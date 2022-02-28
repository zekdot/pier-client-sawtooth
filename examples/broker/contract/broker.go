package contract

import (
	"broker/state"
	"encoding/json"
	"fmt"
	"strconv"
)

const (
	innerMeta = "inner-meta"
	outterMeta = "outter-meta"
	callbackMeta = "callback-meta"
	delimiter = "&"
)

type Broker struct {
	init bool
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

func NewBroker() *Broker {
	return &Broker{
		//init: false,
	}
}

//func (broker *Broker) IsInit() bool {
//	return broker.init
//}

func (broker *Broker)Init(state *state.BrokerState) error {
	fmt.Println("start init")
	inCounter := make(map[string]uint64)
	outCounter := make(map[string]uint64)
	callbackCounter := make(map[string]uint64)

	err := broker.putMap(state, innerMeta, inCounter)
	if err != nil {
		return err
	}
	err = broker.putMap(state, outterMeta, outCounter)
	if err != nil {
		return err
	}
	err = broker.putMap(state, callbackMeta, callbackCounter)
	if err != nil {
		return err
	}
	return nil
}

func (broker *Broker) InterchainDataSwapInvoke(state *state.BrokerState, toId string, contractId string, key string) error {
	//if len
	cid := getChaincodeID()
	newArgs := make([]string, 0)
	newArgs = append(newArgs, toId, cid, contractId, "interchainGet", key, "interchainSet")
	return broker.InterchainInvoke(state, newArgs)
}

func (broker *Broker) InterchainInvoke(state *state.BrokerState, args[] string) error {
	if len(args) < 6 {
		return fmt.Errorf("incorrect number of arguments, expecting 6")
	}
	destChainID := args[0]
	outMeta,err := broker.getMap(state, outterMeta)
	if err != nil {
		return err
	}
	if _, ok := outMeta[destChainID]; !ok {
		outMeta[destChainID] = 0
	}

	tx := &Event{
		Index:         outMeta[destChainID] + 1,
		DstChainID:    destChainID,
		SrcContractID: args[1],
		DstContractID: args[2],
		Func:          args[3],
		Args:          args[4],
		Callback:      args[5],
	}

	outMeta[tx.DstChainID]++
	if err := broker.putMap(state, outterMeta, outMeta); err != nil {
		//return shim.Error(err.Error())
		return err
	}

	txValue, err := json.Marshal(tx)
	if err != nil {
		//return shim.Error(err.Error())
		return err
	}
	key := outMsgKey(tx.DstChainID, strconv.FormatUint(tx.Index, 10))
	if err := state.SetMetaData(key, string(txValue)); err != nil {
		//return shim.Error(fmt.Errorf("persist event: %w", err).Error())
		return err
	}
	return nil
}

func (broker *Broker) PollingEvent(state *state.BrokerState, mStr string)([]*Event, error) {
	m := make(map[string]uint64)
	if err := json.Unmarshal([]byte(mStr), &m); err != nil {
		return nil, err
	}
	outMeta, err := broker.getMap(state, outterMeta)
	if err != nil {
		return nil, err
	}
	events := make([]*Event, 0)
	for addr, idx := range outMeta {
		startPos, ok := m[addr]
		if !ok {
			startPos = 0
		}
		for i := startPos + 1; i <= idx; i ++ {
			eb, err := state.GetMetaData(outMsgKey(addr, strconv.FormatUint(i, 10)))
			if err != nil {
				fmt.Printf("get out event by key %s fail", outMsgKey(addr, strconv.FormatUint(i, 10)))
				continue
			}
			e := &Event{}
			if err := json.Unmarshal([]byte(eb), e); err != nil {
				fmt.Println("unmarshal event fail")
				continue
			}
			events = append(events, e)
		}
	}
	return events, nil
}
package contract

import (
	"broker/state"
	"encoding/json"
	"fmt"
	"strconv"
)

func (broker *Broker)putMap(state *state.BrokerState, metaName string, meta map[string]uint64) error {
	if meta == nil {
		return nil
	}
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	state.SetMetaData(metaName, string(metaBytes))
	return nil
}

func(broker *Broker)getMap(state *state.BrokerState, metaName string) (map[string]uint64, error) {

	meta := make(map[string]uint64)
	metaBytes, err := state.GetMetaData(metaName)
	if err != nil {
		return meta, err
	}

	if err := json.Unmarshal([]byte(metaBytes), &meta); err != nil {
		return nil, err
	}
	return meta, nil
}

func outMsgKey(to string, idx string) string {
	return fmt.Sprintf("out-msg-%s-%s", to, idx)
}

func inMsgKey(from string, idx string) string {
	return fmt.Sprintf("in-msg-%s-%s", from, idx)
}

func (broker *Broker) checkIndex(state *state.BrokerState, addr string, index string, metaName string) error {
	idx, err := strconv.ParseUint(index, 10, 64)
	if err != nil {
		return err
	}
	meta, err := broker.getMap(state, metaName)
	if err != nil {
		return err
	}
	if idx != meta[addr] + 1 {
		return fmt.Errorf("incorrect index, expect %d", meta[addr]+1)
	}
	return nil
}

func getChaincodeID() string {
	return "sawtooth&broker"
}
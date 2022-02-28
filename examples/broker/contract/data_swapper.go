package contract

import (
	"broker/state"
	"fmt"
	"strconv"
	"strings"
)

//func(broker *Broker) Get(state *state.BrokerState, args []string)(string, error) {
//	switch len(args) {
//	case 1:
//		data, err := state.GetData(args[0])
//		if err != nil {
//			return "", err
//		}
//		return data, nil
//	case 3:
//		err := broker.InterchainDataSwapInvoke(state, args[0], args[1], args[2])
//		return "", err
//	default:
//		return "", fmt.Errorf("incorrect number of arguments")
//	}
//}

func (broker *Broker) Set(state *state.BrokerState, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("incorrect number of arguments")
	}
	err := state.SetData(args[0], args[1])
	if err != nil {
		return err
	}
	return nil
}

// get interchain account for transfer contract: setData from,index,tid,name_id,amount
func (broker *Broker) InterchainSet(state *state.BrokerState, args[] string) error {

	if len(args) < 5 {
		return fmt.Errorf("incorrect number of arguments, expecting 5")
	}

	sourceChainID := args[0]
	sequenceNum := args[1]
	targetCID := args[2]
	key := args[3]
	data := args[4]

	if err := broker.checkIndex(state, sourceChainID, sequenceNum, callbackMeta); err != nil {
		return err
	}

	idx, err := strconv.ParseUint(sequenceNum, 10, 64)
	if err != nil {
		return err
	}
	if err := broker.markCallbackCounter(state, sourceChainID, idx); err != nil {
		return err
	}

	splitedCID := strings.Split(targetCID, delimiter)
	if len(splitedCID) != 2 {
		return fmt.Errorf("Target chaincode id %s is not valid", targetCID)
	}

	//b := util.ToChaincodeArgs("interchainSet", key, data)
	//response := ctx.GetStub().InvokeChaincode(splitedCID[1], b, splitedCID[0])
	//if response.Status != shim.OK {
	//	return fmt.Errorf("invoke chaincode '%s' err: %s", splitedCID[1], response.Message)
	//}
	err = state.SetData(key, data)
	if err != nil {
		return err
	}

	return nil
}

// example for calling get: getData from,index,tid,id
func (broker *Broker) InterchainGet(state *state.BrokerState, args[] string) (string, error) {

	if len(args) < 4 {
		return "", fmt.Errorf("incorrect number of arguments, expecting 4")
	}
	sourceChainID := args[0]
	sequenceNum := args[1]
	targetCID := args[2]
	key := args[3]

	if err := broker.checkIndex(state, sourceChainID, sequenceNum, innerMeta); err != nil {
		return "", err
	}

	if err := broker.markInCounter(state, sourceChainID); err != nil {
		return "", err
	}

	splitedCID := strings.Split(targetCID, delimiter)
	if len(splitedCID) != 2 {
		return "", fmt.Errorf("Target chaincode id %s is not valid", targetCID)
	}

	//b := util.ToChaincodeArgs("interchainGet", key)
	//response := ctx.GetStub().InvokeChaincode(splitedCID[1], b, splitedCID[0])
	//if response.Status != shim.OK {
	//	return nil, fmt.Errorf("invoke chaincode '%s' err: %s", splitedCID[1], response.Message)
	//}

	// args[0]: key
	value, err := state.GetData(key)
	if err != nil {
		return "", err
	}

	inKey := inMsgKey(sourceChainID, sequenceNum)
	if err := state.SetMetaData(inKey, value); err != nil {
		return "", err
	}

	return value, nil
}
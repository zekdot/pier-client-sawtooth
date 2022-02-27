package handler

import (
	"broker/contract"
	"broker/payload"
	"broker/state"
	"fmt"
	"github.com/hyperledger/sawtooth-sdk-go/processor"
	"github.com/hyperledger/sawtooth-sdk-go/protobuf/processor_pb2"
)

type BrokerHandler struct {
	broker *contract.Broker
}

func NewHandler(broker *contract.Broker) *BrokerHandler {
	return &BrokerHandler{
		broker: broker,
	}
}

func (handler *BrokerHandler) FamilyName() string {
	return "broker"
}
func(handler *BrokerHandler) FamilyVersions() [] string {
	return []string{"1.0"}
}
func (handler *BrokerHandler) Namespaces()[]string {
	return []string{state.Namespace}
}

func (handler *BrokerHandler) Apply(request *processor_pb2.TpProcessRequest, context *processor.Context) error {
	fmt.Printf("receive %s", string(request.GetPayload()))
	// unmarshal from json bytes
	payload, err := payload.FromBytes(request.GetPayload())
	if err != nil {
		return err
	}
	// TODO not so sure whether cache works
	//fmt.Printf("before context")
	brokerState := state.NewBrokerState(context)
	//fmt.Printf("after context")
	broker := handler.broker
	//fmt.Printf("get function %s", payload.Function)
	// if not init
	fmt.Printf("broker init is %v\n", broker.IsInit())
	if !broker.IsInit() {
		broker.Init(brokerState)
	}
	args := payload.Parameter

	// Sawtooth server only need to finish write operation, read operation is implemented by client
	switch payload.Function {
	case "InterchainDataSwapInvoke":
		return broker.InterchainDataSwapInvoke(brokerState, args[0], args[1], args[2])
	case "InterchainInvoke":
		return broker.InterchainInvoke(brokerState, args)
	case "interchainSet":
		return broker.InterchainSet(brokerState, args)
	case "set":
		//return nil
		return broker.Set(brokerState, args)
	default:
		return &processor.InvalidTransactionError{
			Msg: fmt.Sprintf("Invalid Action : '%v'", payload.Function)}
	}
}
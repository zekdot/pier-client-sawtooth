package handler

import (
	"example/sawtooth-plugin/examples/data_swapper/payload"
	"example/sawtooth-plugin/examples/data_swapper/state"
	"fmt"
	"github.com/hyperledger/sawtooth-sdk-go/processor"
	"github.com/hyperledger/sawtooth-sdk-go/protobuf/processor_pb2"
)

type DSHandler struct {

}

func (self *DSHandler) FamilyName() string {
	return "data_swapper"
}
func(self *DSHandler) FamilyVersions() [] string {
	return []string{"1.0"}
}
func (self *DSHandler) Namespaces()[]string {
	return []string{state.Namespace}
}

func (self *DSHandler) Apply(request *processor_pb2.TpProcessRequest, context *processor.Context) error {
	payload, err := payload.FromBytes(request.GetPayload())
	if err != nil {
		return err
	}
	ds_state := state.NewDSState(context)
	switch payload.Action {
	// 设置合约上的数据
	case "set":
		ds_state.SetData(payload.Key, payload.Value)
		return nil
	default:
		return &processor.InvalidTransactionError{
			Msg: fmt.Sprintf("Invalid Action : '%v'", payload.Action)}
	}
}
package payload

import (
	"github.com/hyperledger/sawtooth-sdk-go/processor"
	"strings"
)

type DSPayload struct {
	Action string	// 动作
	Key string	// key
	Value string	// value
}

func FromBytes(payloadData[] byte) (*DSPayload, error) {
	if payloadData == nil {
		return nil, &processor.InvalidTransactionError{Msg: "Must contain payload"}
	}

	parts := strings.Split(string(payloadData), ",")
	if len(parts) != 3 {
		return nil, &processor.InvalidTransactionError{Msg: "Payload is malformed"}
	}

	payload := DSPayload{}
	payload.Action = parts[0]
	payload.Key = parts[1]

	if len(payload.Key) < 1 {
		return nil, &processor.InvalidTransactionError{Msg: "Key is required"}
	}

	if len(payload.Action) < 1 {
		return nil, &processor.InvalidTransactionError{Msg: "Action is required"}
	}
	if payload.Action == "set" {
		payload.Value = parts[2]
	}
	return &payload, nil
}

package main

import (
	"github.com/meshplus/bitxhub-model/pb"
	"time"
	"fmt"
)

// 生成回调或者回滚的IBTP
func (c *Client) generateCallback(original *pb.IBTP, args [][]byte, proof []byte) (result *pb.IBTP, err error) {
	if original == nil {
		return nil, fmt.Errorf("got nil ibtp to generate receipt: %w", err)
	}
	pd := &pb.Payload{}
	if err := pd.Unmarshal(original.Payload); err != nil {
		return nil, fmt.Errorf("ibtp payload unmarshal: %w", err)
	}
	// 反序列化原来的内容
	originalContent := &pb.Content{}
	if err := originalContent.Unmarshal(pd.Content); err != nil {
		return nil, fmt.Errorf("ibtp payload unmarshal: %w", err)
	}

	content := &pb.Content{
		SrcContractId: originalContent.DstContractId,
		DstContractId: originalContent.SrcContractId,
		Func:          originalContent.Callback,
		Args:          args,
	}
	b, err := content.Marshal()
	if err != nil {
		return nil, err
	}
	retPd := &pb.Payload{
		Content: b,
	}

	pdb, err := retPd.Marshal()
	if err != nil {
		return nil, err
	}
	// 组装完整的IBTP结构
	return &pb.IBTP{
		From:      original.From,
		To:        original.To,
		Index:     original.Index,
		Type:      pb.IBTP_RECEIPT,
		Timestamp: time.Now().UnixNano(),
		Proof:     proof,
		Payload:   pdb,
		Version:   original.Version,
	}, nil
}
package main

import (
	"github.com/meshplus/bitxhub-model/pb"
	"time"
	"fmt"
)

// 生成回调或者回滚的IBTP
func (c *Client) generateCallback(original *pb.IBTP, args [][]byte, proof []byte, status bool) (result *pb.IBTP, err error) {
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
	// 调换一下请求和接受id的位置
	content := &pb.Content{
		SrcContractId: originalContent.DstContractId,
		DstContractId: originalContent.SrcContractId,
	}
	// 如果是执行成功的状态，则写回调的内容
	if status {
		content.Func = originalContent.Callback
		content.Args = append(originalContent.ArgsCb, args...)
	// 否则就写回滚的内容
	} else {
		content.Func = originalContent.Rollback
		content.Args = originalContent.ArgsRb
	}
	// 序列化内容
	b, err := content.Marshal()
	if err != nil {
		return nil, err
	}
	// 加入到载荷之中
	retPd := &pb.Payload{
		Content: b,
	}

	pdb, err := retPd.Marshal()
	if err != nil {
		return nil, err
	}
	// 设置一下状态
	typ := pb.IBTP_RECEIPT_SUCCESS
	if !status {
		typ = pb.IBTP_RECEIPT_FAILURE
	}
	// 组装完整的IBTP结构
	return &pb.IBTP{
		From:      original.From,
		To:        original.To,
		Index:     original.Index,
		Type:      typ,
		Timestamp: time.Now().UnixNano(),
		Proof:     proof,
		Payload:   pdb,
		Version:   original.Version,
	}, nil
}
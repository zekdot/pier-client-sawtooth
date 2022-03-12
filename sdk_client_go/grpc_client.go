package main

import (
	"context"
	"google.golang.org/grpc"
	pb "nju.edu.cn/sdk-client_go/envelope"
	"time"
)

type GrpcClient struct {
	conn *grpc.ClientConn
	client *pb.PostServiceClient
	cancel *context.CancelFunc
	ctx *context.Context
}

func NewGrpcClient(url string) (*GrpcClient, error) {
	conn, err := grpc.Dial(url, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	c := pb.NewPostServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	return &GrpcClient{
		conn: conn,
		client: &c,
		cancel: &cancel,
		ctx: &ctx,
	}, nil
}

func (grpcClient *GrpcClient) close() {
	grpcClient.conn.Close()
	cancel := *grpcClient.cancel
	cancel()
}

func (grpcClient *GrpcClient) get(key string) (string, error) {
	client := *grpcClient.client
	r, err := client.SendEnvelope(*grpcClient.ctx, &pb.EnvelopeRequest{Func: "get", Params: []string{key}})
	if err != nil {
		return "", err
	}
	return r.Result, nil
}

func (grpcClient *GrpcClient) set(key string, value string) error {
	client := *grpcClient.client
	_, err := client.SendEnvelope(*grpcClient.ctx, &pb.EnvelopeRequest{Func: "set", Params: []string{key, value}})
	if err != nil {
		return err
	}
	return nil
}

func (grpcClient *GrpcClient) InterchainGet(targetId string, ccId string, key string) error {
	client := *grpcClient.client
	_, err := client.SendEnvelope(*grpcClient.ctx, &pb.EnvelopeRequest{Func: "interchainGet", Params: []string{targetId, ccId, key}})
	if err != nil {
		return err
	}
	return nil
}

func (grpcClient *GrpcClient) Init() error {
	client := *grpcClient.client
	_, err := client.SendEnvelope(*grpcClient.ctx, &pb.EnvelopeRequest{Func: "init", Params: []string{}})
	if err != nil {
		return err
	}
	return nil
}
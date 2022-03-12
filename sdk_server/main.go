package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	pb "nju.edu.cn/sdk-server/envelope"
)
type Server struct {
	pb.UnimplementedPostServiceServer
	rpcClient *RpcClient
}
func (s *Server) SendEnvelope(ctx context.Context, request *pb.EnvelopeRequest) (*pb.EnvelopeResponse, error) {
	// call rpcClient to execute command on appchain
	switch request.Func {
	case "get":
		log.Printf("get value of " + request.Params[0])
		r, err := s.rpcClient.GetData(request.Params[0])
		if err != nil {
			return nil, err
		}
		//return &pb.EnvelopeResponse{Code: 200, Result: "take a get of " + request.Params[0]}, nil
		//break
		return &pb.EnvelopeResponse{Code: 200, Result: r}, nil
	case "set":
		log.Printf("set " + request.Params[1] + " to " + request.Params[0])
		err := s.rpcClient.SetData(request.Params[0], request.Params[1])
		if err != nil {
			return nil, err
		}
		return nil, nil
		//break
	case "interchainGet":
		log.Printf("interchainGet " + request.Params[2] + " from " + request.Params[0] + " " + request.Params[1])
		//toId, cid, key := request.Params[0], request.Params[1], request.Params[2]
		err := s.rpcClient.InterchainGet(request.Params[0], request.Params[1], request.Params[2])
		if err != nil {
			return nil, err
		}
		return nil, nil
		//break
	case "init":
		log.Printf("init")
		err := s.rpcClient.Init()
		if err != nil {
			return nil, err
		}
		return nil, nil
		//break
	default:
		return nil, fmt.Errorf("no such methods")
	}
	return nil, nil
}
func main()  {
	lis, err := net.Listen("tcp", GRPC_URL)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	rpcClient, err := NewRpcClient(RPC_URL)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	server := &Server{
		rpcClient: rpcClient,
	}
	pb.RegisterPostServiceServer(s, server)
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}

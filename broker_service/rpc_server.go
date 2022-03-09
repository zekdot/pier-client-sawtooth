package main

import (
	"context"
	"fmt"
	"log"
	pb "nju.edu.cn/zekdot/broker_service/envelope"
)

type Server struct {
	broker *BrokerClient
	pb.UnimplementedPostServiceServer
}

func NewServer(client *BrokerClient) *Server {
	return &Server {
		broker: client,
	}
}

func (s *Server) SendEnvelope(ctx context.Context, request *pb.EnvelopeRequest) (*pb.EnvelopeResponse, error) {
	broker := s.broker
	args := request.Params
	switch request.Func {
	case "setValue":
		log.Printf("set %s to %s\n", args[0], args[1])
		err := broker.setValue(args[0], args[1])
		if err != nil {
			//return &pb.EnvelopeResponse{Code: "400", Result: err.Error()}, nil
			return nil, err
		}
		return &pb.EnvelopeResponse{Code: "200", Result: "success"}, nil
	case "getValue":
		log.Printf("get value of %s\n", args[0])
		res, err := broker.getValue(args[0])
		if err != nil {
			//return &pb.EnvelopeResponse{Code: "400", Result: err.Error()}, nil
			return nil, err
		}
		return &pb.EnvelopeResponse{Code: "200", Result: string(res)}, nil
	default:
		return nil, fmt.Errorf("No such method")
	}
	return nil, nil
}
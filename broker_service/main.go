package main

import (
	"google.golang.org/grpc"
	"log"
	"net"
	pb "nju.edu.cn/zekdot/broker_service/envelope"
)

func main() {
	lis, err := net.Listen("tcp", RPC_SERVER_URL)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	brokerClient, _ := NewBrokerClient(SAWTOOTH_URL, KEY_PATH)
	server := NewServer(brokerClient)
	serviceRegister := grpc.NewServer()
	pb.RegisterPostServiceServer(serviceRegister, server)
	log.Printf("server listening at %v", lis.Addr())
	if err := serviceRegister.Serve(lis); err != nil {
		log.Fatal("failed to serve: %v", err)
	}
}
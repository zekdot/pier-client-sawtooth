package main

import "fmt"

type Service struct {
	broker *BrokerClient
}


type ReqArgs struct {
	FuncName string
	Args []string
}

func NewService(broker *BrokerClient) *Service {
	return &Service{
		broker: broker,
	}
}

// send transaction and don't need result
func (s *Service) SetValue(req *ReqArgs, reply *string) error{

	broker := s.broker
	args := req.Args
	fmt.Printf("需要设置%s为%s\n", args[0], args[1])
	err := broker.setValue(args[0], args[1])
	return err
}

// query transaction and need result
func (s *Service) GetValue(req *ReqArgs, reply *string) error{
	broker := s.broker
	args := req.Args
	fmt.Printf("获取%s的值\n", args[0])
	res, err := broker.getValue(args[0])
	*reply = string(res)
	return err
}
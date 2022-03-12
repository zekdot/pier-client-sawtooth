package main

import (
	"fmt"
	"log"
)

func main() {
	client, err := NewGrpcClient("localhost:50051")
	if err != nil {
		log.Fatal(err)
	}
	client.set("key1", "value1")
	value, err := client.get("key1")
	fmt.Println(value)
	//client.InterchainGet("0x78546467877877", "data&swapper", "key2")
	//value, err = client.get("key2")
	//fmt.Println(value)
}
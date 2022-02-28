package main

import "fmt"

func main() {
	brokerClient, _ := NewBrokerClient("http://127.0.0.1:8008", "/home/hzh/.sawtooth/keys/hzh.priv")
	value, err := brokerClient.Get("k1")
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(value)
	outMetaStr, err := brokerClient.getMeta("getOuterMeta")
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(outMetaStr)
	eventStr, err := brokerClient.pollingEvent("{}")
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(eventStr)
}
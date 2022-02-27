package state

import (
	"crypto/sha512"
	"encoding/hex"
	"github.com/hyperledger/sawtooth-sdk-go/processor"
	"strings"
)

var Namespace = hexdigest("broker")[:6]


// 直接存储key到另一个字符串的映射
type BrokerState struct {
	context *processor.Context
	addressCache map[string][]byte
}

func NewBrokerState(context *processor.Context) *BrokerState {
	return &BrokerState{
		context: context,
		addressCache: make(map[string][]byte),
	}
}

func (broker *BrokerState)GetMetaData(key string) (string, error) {
	address := makeAddress(key, "meta")
	// 首先查看缓存
	data, ok := broker.addressCache[address]
	if ok {
		if broker.addressCache[address] != nil {
			return string(data), nil
		}
	}
	// 没有的话再从账本中去取
	results, err := broker.context.GetState([]string{address})
	if err != nil {
		return "", err
	}
	return string(results[address][:]), nil
}

func (broker *BrokerState) SetMetaData(key string, value string) error {
	address := makeAddress(key, "meta")
	data := []byte(value)
	// 进行缓存
	broker.addressCache[address] = data
	// 存储进账本中
	_, err := broker.context.SetState(map[string][] byte {
		address: data,
	})
	return err
}

func (broker *BrokerState)GetData(key string) (string, error) {
	address := makeAddress(key, "regular")
	// 首先查看缓存
	data, ok := broker.addressCache[address]
	if ok {
		if broker.addressCache[address] != nil {
			return string(data), nil
		}
	}
	// 没有的话再从账本中去取
	results, err := broker.context.GetState([]string{address})
	if err != nil {
		return "", err
	}
	return string(results[address][:]), nil
}

func (broker *BrokerState) SetData(key string, value string) error {
	address := makeAddress(key, "regular")
	data := []byte(value)
	// 进行缓存
	broker.addressCache[address] = data
	// 存储进账本中
	_, err := broker.context.SetState(map[string][] byte {
		address: data,
	})
	return err
}

// split regular data and meta
func makeAddress(name string, typeValue string) string {
	return Namespace + hexdigest(typeValue + name)[:64]
}

func hexdigest(str string) string {
	hash := sha512.New()
	hash.Write([]byte(str))
	hashBytes := hash.Sum(nil)
	return strings.ToLower(hex.EncodeToString(hashBytes))
}


package state

import (
	"crypto/sha512"
	"encoding/hex"
	"github.com/hyperledger/sawtooth-sdk-go/processor"
	"strings"
)

var Namespace = hexdigest("data_swapper")[:6]

//type Data struct {
//	key string
//	value string
//}

// 直接存储key到另一个字符串的映射
type DSState struct {
	context *processor.Context
	addressCache map[string][]byte
}

func NewDSState(context *processor.Context) *DSState {
	return &DSState{
		context: context,
		addressCache: make(map[string][]byte),
	}
}

func (self *DSState)GetData(key string) (string, error) {
	address := makeAddress(key)
	// 首先查看缓存
	data, ok := self.addressCache[address]
	if ok {
		if self.addressCache[address] != nil {
			return string(data), nil
		}
	}
	// 没有的话再从账本中去取
	results, err := self.context.GetState([]string{address})
	if err != nil {
		return "", err
	}
	return string(results[address][:]), nil
}

func (self *DSState) SetData(key string, value string) error {
	address := makeAddress(key)
	data := []byte(value)
	// 进行缓存
	self.addressCache[address] = data
	// 存储进账本中
	_, err := self.context.SetState(map[string][] byte {
		address: data,
	})
	return err
}

func makeAddress(name string) string {
	return Namespace + hexdigest(name)[:64]
}

func hexdigest(str string) string {
	hash := sha512.New()
	hash.Write([]byte(str))
	hashBytes := hash.Sum(nil)
	return strings.ToLower(hex.EncodeToString(hashBytes))
}


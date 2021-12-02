package main

import (
	"encoding/json"
	"fmt"
	"os"
	"io/ioutil"
)

func checkFileExists(filename string) (bool) {
	var res = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		res = false
	}
	return res
}

func interface2string(m map[string]interface{}) map[string]string {
	ret := make(map[string]string, len(m))
	for k, v := range m {
		ret[k] = fmt.Sprint(v)
	}
	return ret
}

func interface2uint64(m map[string]interface{}) map[string]uint64 {
	ret := make(map[string]uint64, len(m))
	for k, v := range m {
		ret[k] = uint64(v.(float64))
	}
	return ret
}
// 从文件中读取map
func (c *Client) readMapFromFile(filename string) {
	c.outMeta = make(map[string]uint64)
	c.inMeta = make(map[string]uint64)
	c.callbackMeta = make(map[string]uint64)
	c.inMsgMap = make(map[string]string)
	// 如果不存在这个文件，直接退出
	if !checkFileExists(filename) {

		return
	}
	fmt.Println("文件" + filename + "存在")
	file, _ := os.Open(filename)
	textBytes, _ := ioutil.ReadAll(file)
	fmt.Println("读取到" + string(textBytes))
	// 对读取到的字符串进行反序列化到wholeMap，然后赋值到client上
	wholeMap := make(map[string]interface{})
	json.Unmarshal(textBytes, &wholeMap)
	// fmt.Println(wholeMap)
	c.outMeta = interface2uint64(wholeMap["outMeta"].(map[string]interface{}))
	c.inMeta = interface2uint64(wholeMap["inMeta"].(map[string]interface{}))
	c.callbackMeta = interface2uint64(wholeMap["callbackMeta"].(map[string]interface{}))
	c.inMsgMap = interface2string(wholeMap["inMsgMap"].(map[string]interface{}))
}
// 写入map到文件中
func (c *Client) writeMapToFile(filename string) {
	wholeMap := make(map[string]interface{})
	wholeMap["outMeta"] = c.outMeta
	wholeMap["inMeta"] = c.inMeta
	wholeMap["callbackMeta"] = c.callbackMeta
	wholeMap["inMsgMap"] = c.inMsgMap
	textBytes,_ := json.Marshal(wholeMap)
	// 写入字节数组
	fmt.Println("写入" + string(textBytes))
	err := ioutil.WriteFile(filename, textBytes, 0666);

	if err != nil {
		fmt.Errorf("写入时出现错误")
	}
}
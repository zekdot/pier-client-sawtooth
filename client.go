package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/meshplus/bitxhub-model/pb"
	"github.com/meshplus/pier/pkg/model"
	"github.com/meshplus/pier/pkg/plugins/client"
	"strings"
)

// 这里的定位是sawtooth的一个客户端，因此需要对接sawtooth-sdk的相关接口
// 另外一方面，目前只需要从fabric向sawtooth中查询数据，因此本端暂时不需要主动发起交易，只需要完成接受相关的接口
type Client struct {
	eventC chan *pb.IBTP
	pierId string
	//dsClient *DataSwapperClient	// 用于调用sawtooth交易族的客户端
	outMeta map[string]uint64	// out计数器
	inMeta map[string]uint64	// in计数器
	callbackMeta map[string]uint64	// callback计数器
	inMsgMap map[string]string	// 主动请求的消息
}
var _ client.Client = (*Client)(nil)
// 对方传过来的函数调用
type CallFunc struct {
	Func string   `json:"func"`
	Args [][]byte `json:"args"`
}

// 初始化Plugin服务
func NewClient(configPath, pierId string, extra []byte) (client.Client, error) {
	// 造一个永远不会用的通道
	eventC := make(chan *pb.IBTP)
	c := &Client{}
	// 初始化相关变量和计时器等
	c.eventC = eventC
	c.pierId = pierId
	//c.dsClient, _ = NewDataSwapperClient("http://127.0.0.1:8008", "/home/hzh/.sawtooth/keys/hzh.priv")

	// 设置in、out和callback三个map
	c.outMeta = make(map[string]uint64)
	c.inMeta = make(map[string]uint64)
	c.callbackMeta = make(map[string]uint64)
	c.inMsgMap = make(map[string]string)
	return c, nil
}
// 启动Plugin服务的接口
func (c *Client) Start() error {
	// 啥都不用干接口
	// 采用轮询方式，定时从sawtooth中拉取数据
	//return c.consumer.Start()
	// 从文件中读取出缓存
	c.readMapFromFile("sawtooth.txt")
	return nil
}

// 停止Plugin服务的接口
func (c *Client) Stop() error {
	// 啥都不用干接口
	// 将三个meta和一个msgMap进行持久化存储到文件中
	c.writeMapToFile("sawtooth.txt")
	return nil
}

// Plugin负责将区块链上产生的跨链事件转化为标准的IBTP格式，Pier通过GetIBTP接口获取跨链请求再进行处理
func (c *Client) GetIBTP() chan *pb.IBTP {
	return c.eventC
}

// ToChaincodeArgs converts string args to []byte args
func toChaincodeArgs(args ...string) [][]byte {
	bargs := make([][]byte, len(args))
	for i, arg := range args {
		bargs[i] = []byte(arg)
	}
	return bargs
}

// Plugin 负责执行来源链过来的跨链请求，Pier调用SubmitIBTP提交收到的跨链请求。

// 主要方法，网关需要在这里进行数据的查取
func (c *Client) SubmitIBTP(ibtp *pb.IBTP) (*model.PluginResponse, error) {
	pd := &pb.Payload{}
	ret := &model.PluginResponse{}
	if err := pd.Unmarshal(ibtp.Payload); err != nil {
		return ret, fmt.Errorf("ibtp payload unmarshal: %w", err)
	}
	// content中包含要请求的各种参数等信息
	content := &pb.Content{}
	if err := content.Unmarshal(pd.Content); err != nil {
		return ret, fmt.Errorf("ibtp content unmarshal: %w", err)
	}
	fmt.Println("content.DstContractId ID is " + content.DstContractId)
	fmt.Println("content.Func is " + content.Func)
	fmt.Println("content.Args is ")
	for i, arg := range content.Args {
		fmt.Println(string(i) + string(arg))
	}
	//logger.Info("submit ibtp", "id", ibtp.ID(), "contract", content.DstContractId, "func", content.Func)
	//// 打印接受到的每一个参数
	//for i, arg := range content.Args {
	//	logger.Info("arg", strconv.Itoa(i), string(arg))
	//}
	// 目前只支持interchainGet方法
	if content.Func != "interchainGet" {
		return nil, errors.New("only support interchainGet")
	}



	// 主动发送的消息的回执？？这里应该不会遇到这种情况吧
	//if ibtp.Category() == pb.IBTP_RESPONSE && content.Func == "" {
	//	logger.Info("InvokeIndexUpdate", "ibtp", ibtp.ID())
	//	_, resp, err := c.InvokeIndexUpdate(ibtp.From, ibtp.Index, ibtp.Category())
	//	if err != nil {
	//		return nil, err
	//	}
	//	ret.Status = resp.OK
	//	ret.Message = resp.Message
	//
	//	return ret, nil
	//}
	// 组装IBTP进行提交
	// 调用应用链的相关方法获取到结果
	var result [][]byte
	//var chResp *channel.Response
	callFunc := CallFunc{
		Func: content.Func,
		Args: content.Args,
	}
	bizData, err := json.Marshal(callFunc)
	// 注意下面的区别！一个是InvokeIndexUpdate，另一个是InvokeInterchain，前者仅更新索引，后者则是实际调用InvokeInterchain方法
	if err != nil {
		fmt.Println("Marshal failed?")
		// 如果序列化失败，全部都是失败之后要进行的操作，仅仅更新索引
		ret.Status = false
		ret.Message = fmt.Sprintf("marshal ibtp %s func %s and args: %s", ibtp.ID(), callFunc.Func, err.Error())

		// 简单更新下索引好了，啥都不干
		c.InvokeIndexUpdate(ibtp.From, ibtp.Index)
		if err != nil {
			return nil, err
		}
		//chResp = res
	} else {
		fmt.Println("Marshal success!")
		// 需要调用链码并且获取结果，需要Response结构作为返回值
		resp, err := c.InvokeInterchain(ibtp.From, ibtp.Index, content.DstContractId, bizData)
		if err != nil {
			return nil, fmt.Errorf("invoke interchain for ibtp %s to call %s: %w", ibtp.ID(), content.Func, err)
		}

		ret.Status = resp.OK
		ret.Message = resp.Message

		// if there is callback function, parse returned value
		result = toChaincodeArgs(strings.Split(string(resp.Data), ",")...)
		//chResp = res
	}

	// If is response IBTP, then simply return
	//if ibtp.Category() == pb.IBTP_RESPONSE {
	//	return ret, nil
	//}
	// proof简单来说就是在该链上查到的一个序列号，证明在本链中成功执行了交易，需要从sawtooth的api来找到获取方法，这里我们直接返回一个success的byte数组
	//proof, err := c.getProof(*chResp)
	//if err != nil {
	//	return ret, err
	//}
	proof := []byte("success")

	ret.Result, err = c.generateCallback(ibtp, result, proof)
	fmt.Println("return value")
	if err != nil {
		return nil, err
	}

	return ret, nil


}

// GetOutMessage 负责在跨链合约中查询历史跨链请求。查询键值中to指定目的链，idx指定序号，查询结果为以Plugin负责的区块链作为来源链的跨链请求。
func (c *Client) GetOutMessage(to string, idx uint64) (*pb.IBTP, error) {
	// 这里由于我们使用的本身就是一个set/get合约，所以可以直接获取meta信息
	// 本区块链暂时不可能发出交易
	// =====分析内容如下，首先是to-idx拼接之后读取到的内容====
	// v, err := stub.GetState(key)
	// =====而该内容则是按如下的方式存储的====
	//ccArgs = append(ccArgs, []byte(callFunc.Func))
	//ccArgs = append(ccArgs, callFunc.Args...)
	//response := stub.InvokeChaincode(splitedCID[1], ccArgs, splitedCID[0])
	//if response.Status != shim.OK {
	//	return errorResponse(fmt.Sprintf("invoke chaincode '%s' function %s err: %s", splitedCID[1], callFunc.Func, response.Message))
	//}
	//
	//inKey := broker.inMsgKey(sourceChainID, sequenceNum)
	//value, err := json.Marshal(response)
	//if err != nil {
	//	return errorResponse(err.Error())
	//}
	//if err := stub.PutState(inKey, value);
	// =====解析时则调用了如下的方法====
	// c.unpackIBTP(&response, pb.IBTP_INTERCHAIN)
	// =====unpack的主要逻辑如下：
	//ret := &Event{}
	//if err := json.Unmarshal(response.Payload, ret); err != nil {
	//	return nil, err
	//}
	//proof, err := c.getProof(*response)
	//if err != nil {
	//	return nil, err
	//}
	//ret.Proof = proof
	//return ret.Convert2IBTP(c.pierId, ibtpType), nil

	// =====综上，这里返回的实质是那次请求最后转化出的IBTP报文，所以这里我们最终还是存储IBTP报文
	return nil, nil
}

// GetInMessage 负责在跨链合约中查询历史跨链请求。查询键值中from指定来源链，idx指定序号，查询结果为以Plugin负责的区块链作为目的链的跨链请求。
func (c *Client) GetInMessage(from string, index uint64) ([][]byte, error) {
	// 拿出从本链发出的请求，然后转换为IBTP格式进行返回，这里的话直接把请求转为IBTP，或者直接存IBTP？
	// 最后决定直接存储用逗号隔开的参数字符串
	key := fmt.Sprintf("%s-%s", from, index)
	results := []string{"true"}
	results = append(results, strings.Split(c.inMsgMap[key], ",")...)
	//return toChaincodeArgs(results...), nil
	return toChaincodeArgs(results...), nil
}

// GetInMeta 是获取跨链请求相关的Meta信息的接口。以Plugin负责的区块链为目的链的一系列跨链请求的序号信息。如果Plugin负责A链，则可能有多条链和A进行跨链，如B->A:3; C->A:5。返回的map中，key值为来源链ID，value对应该来源链已发送的最新的跨链请求的序号，如{B:3, C:5}。
func (c *Client) GetInMeta() (map[string]uint64, error) {
	// 直接返回对应的meta
	//return c.unpackMap(response)
	return c.inMeta, nil
}

// GetOutMeta 是获取跨链请求相关的Meta信息的接口。以Plugin负责的区块链为来源链的一系列跨链请求的序号信息。如果Plugin负责A链，则A可能和多条链进行跨链，如A->B:3; A->C:5。返回的map中，key值为目的链ID，value对应已发送到该目的链的最新跨链请求的序号，如{B:3, C:5}。
func (c *Client) GetOutMeta() (map[string]uint64, error) {
	// 直接返回对应的meta
	//return c.unpackMap(response)
	return c.outMeta, nil
}


// GetCallbackMeta 是获取跨链请求相关的Meta信息的接口。以Plugin负责的区块链为来源链的一系列跨链请求的序号信息。如果Plugin负责A链，则A可能和多条链进行跨链，如A->B:3; A->C:5；同时由于跨链请求中支持回调操作，即A->B->A为一次完整的跨链操作，我们需要记录回调请求的序号信息，如A->B->:2; A->C—>A:4。返回的map中，key值为目的链ID，value对应到该目的链最新的带回调跨链请求的序号，如{B:2, C:4}。（注意 CallbackMeta序号可能和outMeta是不一致的，这是由于由A发出的跨链请求部分是没有回调的）
func (c Client) GetCallbackMeta() (map[string]uint64, error) {
	// 直接返回对应的meta
	//return c.unpackMap(response)
	return c.callbackMeta, nil
}

// CommitCallback 执行完IBTP包之后进行一些回调操作
func (c *Client) CommitCallback(ibtp *pb.IBTP) error {
	// 似乎可以不用管，fabric插件中也没有实现
	return nil
}

// Name 描述Plugin负责的区块链的自定义名称，一般和业务相关，如司法链等。
func (c *Client) Name() string {
	return "data_swapper_test"
}

// Type 描述Plugin负责的区块链类型，比如Fabric
func (c *Client) Type() string {
	return "sawtooth"
}

func (c *Client) InvokeInterchain(from string, index uint64, destAddr string, bizCallData []byte) (*Response, error) {
	// 主要的调用方法
	// 目前这里必然是一个请求的回复
	//req := true
	// 直接更新一下索引
	//if req {
		c.inMeta[from] = index
	//} else {
		c.outMeta[from] = index
	//}
	// destAddr应该需要处理一下，这里直接简单修改为"get"
	destAddr = "get"
	// from、index、destAddr等应该是用来进行索引等meta数据的记录的

	// ===========以下是fabric插件中的关键代码，args为全部参数============
	//args := util.ToChaincodeArgs(from, strconv.FormatUint(index, 10), destAddr, req)
	//args = append(args, bizCallData)
	// 链码调用中则是这样的：
	//// 调用的链
	//sourceChainID := args[0]
	//sequenceNum := args[1]
	//// 本链
	//targetCID := args[2]
	//isReq, err := strconv.ParseBool(args[3])
	//if err != nil {
	//	return errorResponse(fmt.Sprintf("cannot parse %s to bool", args[3]))
	//}
	//
	//if err := broker.updateIndex(stub, sourceChainID, sequenceNum, isReq); err != nil {
	//	return errorResponse(err.Error())
	//}
	//
	//splitedCID := strings.Split(targetCID, delimiter)
	//if len(splitedCID) != 2 {
	//	return errorResponse(fmt.Sprintf("Target chaincode id %s is not valid", targetCID))
	//}
	//
	//callFunc := &CallFunc{}
	//if err := json.Unmarshal([]byte(args[4]), callFunc); err != nil {
	//	return errorResponse(fmt.Sprintf("unmarshal call func failed for %s", args[4]))
	//}
	// ===============================================


	// 进行实际的链码调用，bizCallData为请求参数，这里是从CallBackFunc序列化得来的，这里直接给他反序列化了
	bizCallFun := &CallFunc{}
	json.Unmarshal(bizCallData, bizCallFun)
	//key := string(bizCallFun.Args[0][:])
	value := "sawtooth_result"
	// 调用sawtooth客户端来得到调用结果，事实上目前只会调用get方法，所以只需要得到get的key参数即可，key参数为简单转换为字节数组数组的字符串数组，所以只需要定位key的索引然后字节数组转字符串即可
	response := &Response{
		OK:   true,
		Data: []byte(value),
	}
	//if err := json.Unmarshal(res.Payload, response); err != nil {
	//	return nil, err
	//}

	return response, nil
}

func (c *Client) IncreaseInMeta(original *pb.IBTP) (*pb.IBTP, error) {
	// 直接更新索引
	c.InvokeIndexUpdate(original.From, original.Index)
	proof := []byte("success")
	ibtp, err := c.generateCallback(original, nil, proof)
	if err != nil {
		return nil, err
	}
	return ibtp, nil
}

// 该方法似乎只更新了索引，返回索引更新的结果
func (c Client) InvokeIndexUpdate(from string, index uint64) ( *Response, error) {
	//
	//req := true
	// 直接更新一下索引
	//if req {
		c.inMeta[from] = index
	//} else {
		c.outMeta[from] = index
	//}
	//// 构造请求参数
	//args := util.ToChaincodeArgs(from, strconv.FormatUint(index, 10), req)
	//request := channel.Request{
	//	ChaincodeID: c.meta.CCID,
	//	Fcn:         InvokeIndexUpdateMethod,
	//	Args:        args,
	//}
	//// 进行实际的请求
	//res, err := c.consumer.ChannelClient.Execute(request)
	//if err != nil {
	//	return nil, nil, err
	//}
	//// 返回请求的结果
	//response := &Response{}
	//if err := json.Unmarshal(res.Payload, response); err != nil {
	//	return nil, nil, err
	//}

	//return &res, response, nil
	return nil, nil
}

type Response struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
	Data    []byte `json:"data"`
}
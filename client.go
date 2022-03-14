package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Rican7/retry"
	"github.com/Rican7/retry/strategy"

	//"github.com/hyperledger/fabric-protos-go/common"
	//"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"

	"github.com/meshplus/bitxhub-kit/log"
	"github.com/meshplus/bitxhub-model/pb"
	"github.com/meshplus/pier/pkg/model"
	"github.com/meshplus/pier/pkg/plugins/client"
)

var logger = log.NewWithModule("client")

var _ client.Client = (*Client)(nil)

const (

	PollingEventMethod    = "pollingEvent"
<<<<<<< HEAD
	FabricType            = "sawtooth"
=======
	//FabricType            = "fabric2.0"
	SawtoothType		= "sawtooth"
>>>>>>> heavy-plugin
)

type Client struct {
	//meta     *ContractMeta
	//consumer *Consumer
	eventC   chan *pb.IBTP
	pierId   string
	name     string
	outMeta  map[string]uint64
	ticker   *time.Ticker
	done     chan bool
	client   *RpcClient
}

func NewClient(configPath, pierId string, extra []byte) (client.Client, error) {
	// read config of fabric
	//fabricConfig, err := UnmarshalConfig(configPath)

	//if err != nil {
	//	return nil, err
	//}

	eventC := make(chan *pb.IBTP)

	m := make(map[string]uint64)
	if err := json.Unmarshal(extra, &m); err != nil {
		return nil, fmt.Errorf("unmarshal extra for plugin :%w", err)
	}
	if m == nil {
		m = make(map[string]uint64)
	}

	done := make(chan bool)

	//rpcClient, err := rpc.DialHTTP("tcp", RPC_URL)
	rpcClient, err := NewRpcClient(RPC_URL)
	if err != nil {
		logger.Fatal("dialing: ", err)
	}

	return &Client{
		//consumer: csm,
		eventC:   eventC,
		//meta:     c,
		pierId:   pierId,
		name:     "sawtooth",// fabricConfig.Name,
		outMeta:  m,
		ticker:   time.NewTicker(2 * time.Second),
		done:     done,
		client: rpcClient,
	}, nil
}

func (c *Client) Start() error {
	logger.Info("Fabric consumer started")
	go c.polling()
	//return c.consumer.Start()
	return nil
}

// polling event from broker
func (c *Client) polling() {
	for {
		select {
		case <-c.ticker.C:
			evs, err := c.client.Polling(c.outMeta)
			if err != nil {
				return
			}
			for _, ev := range evs {
				//ev.Proof = proof
				c.eventC <- ev.Convert2IBTP(c.pierId, pb.IBTP_INTERCHAIN)
				if c.outMeta == nil {
					c.outMeta = make(map[string]uint64)
				}
				c.outMeta[ev.DstChainID]++
			}
		case <-c.done:
			logger.Info("Stop long polling")
			return
		}
	}
}

func (c *Client) Stop() error {
	c.ticker.Stop()
	c.done <- true
	//return c.consumer.Shutdown()
	return nil
}

func (c *Client) Name() string {
	return c.name
}

func (c *Client) Type() string {
	return SawtoothType
}

func (c *Client) GetIBTP() chan *pb.IBTP {
	return c.eventC
}

func (c *Client) SubmitIBTP(ibtp *pb.IBTP) (*model.PluginResponse, error) {
	pd := &pb.Payload{}
	ret := &model.PluginResponse{}
	if err := pd.Unmarshal(ibtp.Payload); err != nil {
		return ret, fmt.Errorf("ibtp payload unmarshal: %w", err)
	}
	content := &pb.Content{}
	if err := content.Unmarshal(pd.Content); err != nil {
		return ret, fmt.Errorf("ibtp content unmarshal: %w", err)
	}

	//args := make([]string, len(content.Args) + 3)
	//
	//args = append(args, ibtp.From, strconv.FormatUint(ibtp.Index, 10), content.DstContractId)
	// these three parameters are sure
	args := []string {ibtp.From, strconv.FormatUint(ibtp.Index, 10), content.DstContractId}
	// others are not sure
	for i := 0; i < len(content.Args); i ++ {
		args = append(args, string(content.Args[i]))
	}

	// DEBUG
	fmt.Println("whole args is ")
	for i, arg := range args {
		fmt.Println(string(i) + arg + "!")
	}
	//funcName := toPublicFunction(content.Func)

	fmt.Println("funcName is ", content.Func)
	//args = append(args, content.Args...)
	//request := channel.Request{
	//	ChaincodeID: c.meta.CCID,
	//	Fcn:         content.Func,
	//	Args:        args,
	//}

	// retry executing
	var res string
	var proof []byte
	var err error
	if err := retry.Retry(func(attempt uint) error {
		if content.Func == "interchainGet" {
			res, _ = c.client.InterchainGet(args)
		} else if content.Func == "interchainSet" {
			c.client.InterchainSet(args)
		}
		//res, err = c.contract.SubmitTransaction(funcName, args...)
		//res, err = c.consumer.ChannelClient.Execute(request)
		if err != nil {
			return fmt.Errorf("execute request: %w", err)
		}

		return nil
	}, strategy.Wait(2*time.Second)); err != nil {
		logger.Panicf("Can't send rollback ibtp back to bitxhub: %s", err.Error())
	}

	//response := &Response{}
	//if err := json.Unmarshal(res.Payload, response); err != nil {
	//	return nil, err
	//}

	// if there is callback function, parse returned value
	result := toChaincodeArgs(strings.Split(res, ",")...)
	newArgs := make([][]byte, 0)
	ret.Status = true// response.OK
	ret.Message = "success"// response.Message

	// DEBUG
	fmt.Println("content.DstContractId ID is " + content.DstContractId)
	fmt.Println("content.Func is " + content.Func)
	fmt.Println("content.Args is ")
	for i, arg := range content.Args {
		fmt.Println(string(i) + string(arg) + "!")
	}

	// If no callback function to invoke, then simply return
	if content.Callback == "" {
		return ret, nil
	}

	//proof, err = c.getProof(res)
	proof = []byte("success")
	//if err != nil {
	//	return ret, err
	//}

	switch content.Func {
	case "interchainGet":
		newArgs = append(newArgs, content.Args[0])
		newArgs = append(newArgs, result...)
		//case "interchainCharge":
		//	newArgs = append(newArgs, []byte(strconv.FormatBool(response.OK)), content.Args[0])
		//	newArgs = append(newArgs, content.Args[2:]...)
	}

	ret.Result, err = c.generateCallback(ibtp, newArgs, proof)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (c *Client) GetOutMessage(to string, idx uint64) (*pb.IBTP, error) {
	ret, err := c.client.GetOutMessage(to, idx)
	if err != nil {
		return nil, err
	}
	return ret.Convert2IBTP(c.pierId, pb.IBTP_INTERCHAIN), nil
}

func (c *Client) GetInMessage(from string, idx uint64) ([][]byte, error) {
	return c.client.GetInMessage(from, idx)
}

func (c *Client) GetInMeta() (map[string]uint64, error) {
	return c.client.GetInnerMeta()
}

func (c *Client) GetOutMeta() (map[string]uint64, error) {
	return c.client.GetOuterMeta()
}

func (c Client) GetCallbackMeta() (map[string]uint64, error) {
	return c.client.GetCallbackMeta()
}

func (c *Client) CommitCallback(ibtp *pb.IBTP) error {
	return nil
}

func (c *Client) unpackIBTP(response *channel.Response, ibtpType pb.IBTP_Type) (*pb.IBTP, error) {
	ret := &Event{}
	if err := json.Unmarshal(response.Payload, ret); err != nil {
		return nil, err
	}

	return ret.Convert2IBTP(c.pierId, ibtpType), nil
}

func (c *Client) unpackMap(response channel.Response) (map[string]uint64, error) {
	if response.Payload == nil {
		return nil, nil
	}
	r := make(map[string]uint64)
	err := json.Unmarshal(response.Payload, &r)
	if err != nil {
		return nil, fmt.Errorf("unmarshal payload :%w", err)
	}

	return r, nil
}


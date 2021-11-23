package main
import (
	"encoding/json"
	"fmt"
	"github.com/meshplus/bitxhub-model/pb"
	"plugin"
	"strconv"
	"strings"
)

// 这里的定位是sawtooth的一个客户端，因此需要对接sawtooth-sdk的相关接口
type Client struct {
	eventC chan *pb.IBTP
	pierId string
	name string
}

// 初始化Plugin服务
func (c *Client) Initialize(configPath, pierId string, extra []byte) error {

	// 初始化相关变量和计时器等
	return nil
}
// 启动Plugin服务的接口
func (c *Client) Start() error {

	// 采用轮询方式，定时从sawtooth中拉取数据
	return c.consumer.Start()
}

// 停止Plugin服务的接口
func (c *Client) Stop() error {
	// 停止计时器的工作
}

// Plugin负责将区块链上产生的跨链事件转化为标准的IBTP格式，Pier通过GetIBTP接口获取跨链请求再进行处理
func (c *Client) GetIBTP() chan *pb.IBTP {
	return c.eventC
}

// Plugin 负责执行来源链过来的跨链请求，Pier调用SubmitIBTP提交收到的跨链请求。
func (c *Client) SubmitIBTP(ibtp *pb.IBTP) (*pb.SubmitIBTPResponse, error) {
	// 组装IBTP进行提交
	return ret, nil
}

// GetOutMessage 负责在跨链合约中查询历史跨链请求。查询键值中to指定目的链，idx指定序号，查询结果为以Plugin负责的区块链作为来源链的跨链请求。
func (c *Client) GetOutMessage(to string, idx uint64) (*pb.IBTP, error) {
	// 从链上取得outMeta对应的信息
}

// GetInMessage 负责在跨链合约中查询历史跨链请求。查询键值中from指定来源链，idx指定序号，查询结果为以Plugin负责的区块链作为目的链的跨链请求。
func (c *Client) GetInMessage(from string, index uint64) ([][]byte, error) {

	return util.ToChaincodeArgs(results...), nil
}

// GetInMeta 是获取跨链请求相关的Meta信息的接口。以Plugin负责的区块链为目的链的一系列跨链请求的序号信息。如果Plugin负责A链，则可能有多条链和A进行跨链，如B->A:3; C->A:5。返回的map中，key值为来源链ID，value对应该来源链已发送的最新的跨链请求的序号，如{B:3, C:5}。
func (c *Client) GetInMeta() (map[string]uint64, error) {

	return c.unpackMap(response)
}

// GetOutMeta 是获取跨链请求相关的Meta信息的接口。以Plugin负责的区块链为来源链的一系列跨链请求的序号信息。如果Plugin负责A链，则A可能和多条链进行跨链，如A->B:3; A->C:5。返回的map中，key值为目的链ID，value对应已发送到该目的链的最新跨链请求的序号，如{B:3, C:5}。
func (c *Client) GetOutMeta() (map[string]uint64, error) {

	return c.unpackMap(response)
}

// GetCallbackMeta 是获取跨链请求相关的Meta信息的接口。以Plugin负责的区块链为来源链的一系列跨链请求的序号信息。如果Plugin负责A链，则A可能和多条链进行跨链，如A->B:3; A->C:5；同时由于跨链请求中支持回调操作，即A->B->A为一次完整的跨链操作，我们需要记录回调请求的序号信息，如A->B->:2; A->C—>A:4。返回的map中，key值为目的链ID，value对应到该目的链最新的带回调跨链请求的序号，如{B:2, C:4}。（注意 CallbackMeta序号可能和outMeta是不一致的，这是由于由A发出的跨链请求部分是没有回调的）
func (c Client) GetCallbackMeta() (map[string]uint64, error) {

	return c.unpackMap(response)
}

// CommitCallback 执行完IBTP包之后进行一些回调操作
func (c *Client) CommitCallback(ibtp *pb.IBTP) error {
	return nil
}

// GetReceipt 获取一个已被执行IBTP的回执
func (c *Client) GetReceipt(ibtp *pb.IBTP) (*pb.IBTP, error) {

	return c.generateCallback(ibtp, result[1:], nil, status)
}

// Name 描述Plugin负责的区块链的自定义名称，一般和业务相关，如司法链等。
func (c *Client) Name() string {
	return c.name
}

// Type 描述Plugin负责的区块链类型，比如Fabric
func (c *Client) Type() string {
	return FabricType
}


// polling event from broker
func (c *Client) polling() {

}

func (c *Client) getProof(response channel.Response) ([]byte, error) {

	return proof, nil
}

func (c *Client) InvokeInterchain(from string, index uint64, destAddr string, category pb.IBTP_Category, bizCallData []byte) (*channel.Response, *Response, error) {

}







// @ibtp is the original ibtp merged from this appchain
func (c *Client) RollbackIBTP(ibtp *pb.IBTP, isSrcChain bool) (*pb.RollbackIBTPResponse, error) {

	return ret, nil
}

func (c *Client) IncreaseInMeta(original *pb.IBTP) (*pb.IBTP, error) {

	return ibtp, nil
}



func (c Client) InvokeIndexUpdate(from string, index uint64, category pb.IBTP_Category) (*channel.Response, *Response, error) {

	return &res, response, nil
}

func (c *Client) unpackIBTP(response *channel.Response, ibtpType pb.IBTP_Type) (*pb.IBTP, error) {

	return ret.Convert2IBTP(c.pierId, ibtpType), nil
}

func (c *Client) unpackMap(response channel.Response) (map[string]uint64, error) {

	return r, nil
}

type handler struct {
	eventFilter string
	eventC      chan *pb.IBTP
	ID          string
}

func newFabricHandler(eventFilter string, eventC chan *pb.IBTP, pierId string) (*handler, error) {
	return &handler{
		eventC:      eventC,
		eventFilter: eventFilter,
		ID:          pierId,
	}, nil
}

func (h *handler) HandleMessage(deliveries *fab.CCEvent, payload []byte) {
	if deliveries.EventName == h.eventFilter {
		e := &pb.IBTP{}
		if err := e.Unmarshal(deliveries.Payload); err != nil {
			return
		}
		e.Proof = payload

		h.eventC <- e
	}
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: plugins.Handshake,
		Plugins: map[string]plugin.Plugin{
			plugins.PluginName: &plugins.AppchainGRPCPlugin{Impl: &Client{}},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})

	logger.Info("Plugin server down")
}
package main

import (
	bytes2 "bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/sawtooth-sdk-go/protobuf/batch_pb2"
	"github.com/hyperledger/sawtooth-sdk-go/protobuf/transaction_pb2"
	"github.com/hyperledger/sawtooth-sdk-go/signing"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type BrokerClient struct {
	url string
	signer *signing.Signer
}

func NewBrokerClient(url string, keyfile string) (*BrokerClient, error) {
	var privateKey signing.PrivateKey
	if keyfile != "" {
		// Read private key file
		privateKeyStr, err := ioutil.ReadFile(keyfile)
		if err != nil {
			return &BrokerClient{},
				errors.New(fmt.Sprintf("Failed to read private key: %v", err))
		}
		// Get private key object
		privateKey = signing.NewSecp256k1PrivateKey(privateKeyStr)
	} else {
		privateKey = signing.NewSecp256k1Context().NewRandomPrivateKey()
	}
	cryptoFactory := signing.NewCryptoFactory(signing.NewSecp256k1Context())
	signer := cryptoFactory.NewSigner(privateKey)
	return &BrokerClient{url, signer}, nil
}

func (broker *BrokerClient) Init(wait uint) (string, error) {
	rand.Seed(time.Now().Unix())
	// construct the payload information in CBOR format
	payloadData := make(map[string]interface{})
	payloadData["Function"] = "init"
	args := make([]string, 0)
	//payloadData["Key"] = [key, value]
	payloadData["Parameter"] = append(args, strconv.Itoa(rand.Int()))
	//payload := fmt.Sprintf("%s,%s,%s", payloadData["Action"], payloadData["Key"], payloadData["Value"])
	payload, err := json.Marshal(payloadData)
	if err != nil {
		return "", err
	}
	// construct the address
	address := broker.getAddress("inner-meta", "meta")
	fmt.Println("local calc address is " + address)
	fmt.Printf("send %v\n", string(payload))
	//fmt.Printf("get address hash %v\n", address)
	// Construct TransactionHeader
	rawTransactionHeader := transaction_pb2.TransactionHeader{
		SignerPublicKey:  broker.signer.GetPublicKey().AsHex(),
		FamilyName:       FAMILY_NAME,
		FamilyVersion:    FAMILY_VERSION,
		Dependencies:     []string{}, // empty dependency list
		Nonce:            strconv.Itoa(rand.Int()),
		BatcherPublicKey: broker.signer.GetPublicKey().AsHex(),
		Inputs:           []string{
			broker.getAddress("inner-meta", "meta"),
			broker.getAddress("outter-meta", "meta"),
			broker.getAddress("callback-meta", "meta")},
		Outputs:          []string{
			broker.getAddress("inner-meta", "meta"),
			broker.getAddress("outter-meta", "meta"),
			broker.getAddress("callback-meta", "meta")},
		PayloadSha512:    Sha512HashValue(string(payload)),
	}
	transactionHeader, err := proto.Marshal(&rawTransactionHeader)
	if err != nil {
		return "", errors.New(
			fmt.Sprintf("Unable to serialize transaction header: %v", err))
	}

	// Signature of TransactionHeader
	transactionHeaderSignature := hex.EncodeToString(
		broker.signer.Sign(transactionHeader))

	// Construct Transaction
	transaction := transaction_pb2.Transaction{
		Header:          transactionHeader,
		HeaderSignature: transactionHeaderSignature,
		Payload:         []byte(payload),
	}

	// Get BatchList
	rawBatchList, err := broker.createBatchList(
		[]*transaction_pb2.Transaction{&transaction})
	if err != nil {
		return "", errors.New(
			fmt.Sprintf("Unable to construct batch list: %v", err))
	}
	batchId := rawBatchList.Batches[0].HeaderSignature
	batchList, err := proto.Marshal(&rawBatchList)
	if err != nil {
		return "", errors.New(
			fmt.Sprintf("Unable to serialize batch list: %v", err))
	}

	if wait > 0 {
		waitTime := uint(0)
		startTime := time.Now()
		response, err := broker.sendRequest(
			BATCH_SUBMIT_API, batchList, CONTENT_TYPE_OCTET_STREAM, "init")
		if err != nil {
			return "", err
		}
		for waitTime < wait {
			status, err := broker.getStatus(batchId, wait-waitTime)
			if err != nil {
				return "", err
			}
			waitTime = uint(time.Now().Sub(startTime))
			if status != "PENDING" {
				return response, nil
			}
		}
		return response, nil
	}

	return broker.sendRequest(
		BATCH_SUBMIT_API, batchList, CONTENT_TYPE_OCTET_STREAM, "init")
}

func (broker *BrokerClient) Set(
	key string, value string, wait uint) (string, error) {
	return broker.sendTransaction("set", key, value, wait)
}

func (broker *BrokerClient) Get(
	name string) (string, error) {
	apiSuffix := fmt.Sprintf("%s/%s", STATE_API, broker.getAddress(name, "regular"))
	fmt.Printf("apiSuffix is %s\n", apiSuffix)
	response, err := broker.sendRequest(apiSuffix, []byte{}, "", name)
	if err != nil {
		return "", err
	}
	responseMap := make(map[interface{}]interface{})

	err = yaml.Unmarshal([]byte(response), &responseMap)
	if err != nil {
		return "", errors.New(fmt.Sprint("Error reading response: %v", err))
	}
	data, ok := responseMap["data"].(string)
	if !ok {
		return "", errors.New("Error reading as string")
	}
	responseData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", errors.New(fmt.Sprint("Error decoding response: %v", err))
	}
	return fmt.Sprintf("%v", string(responseData[:])), nil
}

func (broker *BrokerClient) sendRequest(

	apiSuffix string,
	data []byte,
	contentType string,
	name string) (string, error) {

	// Construct URL
	var url string
	if strings.HasPrefix(broker.url, "http://") {
		url = fmt.Sprintf("%s/%s", broker.url, apiSuffix)
	} else {
		url = fmt.Sprintf("http://%s/%s", broker.url, apiSuffix)
	}

	// Send request to validator URL
	var response *http.Response
	var err error
	if len(data) > 0 {
		response, err = http.Post(url, contentType, bytes2.NewBuffer(data))
	} else {
		response, err = http.Get(url)
	}
	if err != nil {
		return "", errors.New(
			fmt.Sprintf("Failed to connect to REST API: %v", err))
	}
	if response.StatusCode == 404 {
		logger.Debug(fmt.Sprintf("%v", response))
		return "", errors.New(fmt.Sprintf("No such key: %s", name))
	} else if response.StatusCode >= 400 {
		return "", errors.New(
			fmt.Sprintf("Error %d: %s", response.StatusCode, response.Status))
	}
	defer response.Body.Close()
	reponseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error reading response: %v", err))
	}
	return string(reponseBody), nil
}

func (broker *BrokerClient) getStatus(
	batchId string, wait uint) (string, error) {

	// API to call
	apiSuffix := fmt.Sprintf("%s?id=%s&wait=%d",
		BATCH_STATUS_API, batchId, wait)
	response, err := broker.sendRequest(apiSuffix, []byte{}, "", "")
	if err != nil {
		return "", err
	}

	responseMap := make(map[interface{}]interface{})
	err = yaml.Unmarshal([]byte(response), &responseMap)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error reading response: %v", err))
	}
	entry :=
		responseMap["data"].([]interface{})[0].(map[interface{}]interface{})
	return fmt.Sprint(entry["status"]), nil
}

func (broker *BrokerClient) sendTransaction(
	function string, key string, value string, wait uint) (string, error) {
	rand.Seed(time.Now().Unix())
	// construct the payload information in CBOR format
	payloadData := make(map[string]interface{})
	payloadData["Function"] = function
	args := make([]string, 0)
	// don't know why
	args = append(args, key, value, strconv.Itoa(rand.Int()))
	//payloadData["Key"] = [key, value]
	payloadData["Parameter"] = args
	//payload := fmt.Sprintf("%s,%s,%s", payloadData["Action"], payloadData["Key"], payloadData["Value"])
	payload, err := json.Marshal(payloadData)
	if err != nil {
		return "", err
	}
	// construct the address
	address := broker.getAddress(key, "regular")
	fmt.Printf("send %v\n", string(payload))
	fmt.Printf("get address hash %v\n", address)
	// Construct TransactionHeader
	rawTransactionHeader := transaction_pb2.TransactionHeader{
		SignerPublicKey:  broker.signer.GetPublicKey().AsHex(),
		FamilyName:       FAMILY_NAME,
		FamilyVersion:    FAMILY_VERSION,
		Dependencies:     []string{}, // empty dependency list
		Nonce:            strconv.Itoa(rand.Int()),
		BatcherPublicKey: broker.signer.GetPublicKey().AsHex(),
		Inputs:           []string{address},
		Outputs:          []string{address},
		PayloadSha512:    Sha512HashValue(string(payload)),
	}
	transactionHeader, err := proto.Marshal(&rawTransactionHeader)
	if err != nil {
		return "", errors.New(
			fmt.Sprintf("Unable to serialize transaction header: %v", err))
	}

	// Signature of TransactionHeader
	transactionHeaderSignature := hex.EncodeToString(
		broker.signer.Sign(transactionHeader))

	// Construct Transaction
	transaction := transaction_pb2.Transaction{
		Header:          transactionHeader,
		HeaderSignature: transactionHeaderSignature,
		Payload:         []byte(payload),
	}

	// Get BatchList
	rawBatchList, err := broker.createBatchList(
		[]*transaction_pb2.Transaction{&transaction})
	if err != nil {
		return "", errors.New(
			fmt.Sprintf("Unable to construct batch list: %v", err))
	}
	batchId := rawBatchList.Batches[0].HeaderSignature
	batchList, err := proto.Marshal(&rawBatchList)
	if err != nil {
		return "", errors.New(
			fmt.Sprintf("Unable to serialize batch list: %v", err))
	}

	if wait > 0 {
		waitTime := uint(0)
		startTime := time.Now()
		response, err := broker.sendRequest(
			BATCH_SUBMIT_API, batchList, CONTENT_TYPE_OCTET_STREAM, key)
		if err != nil {
			return "", err
		}
		for waitTime < wait {
			status, err := broker.getStatus(batchId, wait-waitTime)
			if err != nil {
				return "", err
			}
			waitTime = uint(time.Now().Sub(startTime))
			if status != "PENDING" {
				return response, nil
			}
		}
		return response, nil
	}

	return broker.sendRequest(
		BATCH_SUBMIT_API, batchList, CONTENT_TYPE_OCTET_STREAM, key)
}

func (broker *BrokerClient) getPrefix() string {
	return Sha512HashValue(FAMILY_NAME)[:FAMILY_NAMESPACE_ADDRESS_LENGTH]
}

func (broker *BrokerClient) getAddress(name string, typeValue string) string {
	prefix := broker.getPrefix()
	nameAddress := Sha512HashValue(typeValue + name)[:FAMILY_VERB_ADDRESS_LENGTH]
	return prefix + nameAddress
}

func (broker *BrokerClient) createBatchList(
	transactions []*transaction_pb2.Transaction) (batch_pb2.BatchList, error) {

	// Get list of TransactionHeader signatures
	transactionSignatures := []string{}
	for _, transaction := range transactions {
		transactionSignatures =
			append(transactionSignatures, transaction.HeaderSignature)
	}

	// Construct BatchHeader
	rawBatchHeader := batch_pb2.BatchHeader{
		SignerPublicKey: broker.signer.GetPublicKey().AsHex(),
		TransactionIds:  transactionSignatures,
	}
	batchHeader, err := proto.Marshal(&rawBatchHeader)
	if err != nil {
		return batch_pb2.BatchList{}, errors.New(
			fmt.Sprintf("Unable to serialize batch header: %v", err))
	}

	// Signature of BatchHeader
	batchHeaderSignature := hex.EncodeToString(
		broker.signer.Sign(batchHeader))

	// Construct Batch
	batch := batch_pb2.Batch{
		Header:          batchHeader,
		Transactions:    transactions,
		HeaderSignature: batchHeaderSignature,
	}

	// Construct BatchList
	return batch_pb2.BatchList{
		Batches: []*batch_pb2.Batch{&batch},
	}, nil
}
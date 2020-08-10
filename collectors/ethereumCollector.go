package collectors

import (
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

type EthereumCollector struct{}

func (e EthereumCollector) sourceName() string {
	return "ethereum"
}

func (e EthereumCollector) collectEvent() (string, string) {
	sources := []string{"localNode", "infura", "etherscan", "rivet"}
	for _, source := range sources {
		blockHash, blockNumber, valid := getLastBlock(source)
		if valid {
			return blockHash, blockNumber
		}
	}
	log.Error("no ethereum source available")
	return "", ""
}

func getEthToken(source string) string {
	configJsonFile, err := os.Open("collectors/ethereumConfig.json")
	if err != nil {
		fmt.Println(err)
	}
	defer configJsonFile.Close()
	ethTokens := make(map[string]string)
	byteValue, _ := ioutil.ReadAll(configJsonFile)
	json.Unmarshal(byteValue, &ethTokens)

	return ethTokens[source]
}

func getLastBlock(source string) (string, string, bool) {
	var ethAPI string
	switch source {
	case "localNode":
		ethAPI = "https://eth.labs.clcert.cl%s"
	case "infura":
		ethAPI = "https://mainnet.infura.io/v3/%s"
	case "etherscan":
		ethAPI = "https://api.etherscan.io/api?module=proxy&action=eth_getBlockByNumber&tag=latest&boolean=false&apikey=%s"
	case "rivet":
		ethAPI = "https://%s.eth.rpc.rivet.cloud/"
	default:
		ethAPI = ""
	}

	jsonStr := []byte(`{"jsonrpc": "2.0", "method": "eth_getBlockByNumber", "id": "1", "params": ["latest", false]}`)
	resp, err := http.Post(fmt.Sprintf(ethAPI, getEthToken(source)), "application/json", bytes.NewReader(jsonStr))

	if err != nil {
		log.WithFields(log.Fields{
			"ethSource": source,
		}).Error("failed to get ethereum event")
		return "", "", false
	}

	if resp.StatusCode != 200 {
		log.WithFields(log.Fields{
			"ethSource": source,
		}).Error("ethereum response error, status code: " + strconv.Itoa(resp.StatusCode))
		return "", "", false
	}

	body := resp.Body
	defer body.Close()

	response, _ := ioutil.ReadAll(body)
	blockInfo := make(map[string]map[string]string)
	_ = json.Unmarshal(response, &blockInfo)
	if _, ok := blockInfo["error"]; ok {
		log.WithFields(log.Fields{
			"ethSource": source,
		}).Error("ethereum response with error")
		log.Error(blockInfo["error"])
		return "", "", false
	} else {
		var lastBlockHash string
		lastBlockNumber := blockInfo["result"]["number"][2:]
		if isEven(lastBlockNumber) {
			lastBlockHash = blockInfo["result"]["hash"][2:]
		} else {
			lastBlockHash = blockInfo["result"]["parentHash"][2:]
			lastBlockNumber = subtractOne(lastBlockNumber)
		}
		return lastBlockHash, lastBlockNumber, true
	}
}

func isEven(hexNumber string) bool {
	num, _ := strconv.ParseInt(hexNumber, 16, 64)
	return num%2 == 0
}

func subtractOne(hexNumber string) string {
	num, _ := strconv.ParseInt(hexNumber, 16, 64)
	output := num - 1
	return fmt.Sprintf("%x", output)
}

func (e EthereumCollector) estimateEntropy() int {
	return 0
}

func (e EthereumCollector) getCanonicalForm(s string) string {
	return s
}

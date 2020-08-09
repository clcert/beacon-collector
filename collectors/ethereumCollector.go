package collectors

import (
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strconv"
)

type EthereumCollector struct{}

func (e EthereumCollector) sourceName() string {
	return "ethereum"
}

func (e EthereumCollector) collectEvent() (string, string) {
	ethAPI := "https://eth.labs.clcert.cl"
	jsonStr := []byte(`{"jsonrpc": "2.0", "method": "eth_getBlockByNumber", "id": "1", "params": ["latest", false]}`)
	resp, err := http.Post(ethAPI, "application/json", bytes.NewReader(jsonStr))

	if err != nil {
		log.Error("failed to get ethereum event")
		return "", ""
	}

	if resp.StatusCode != 200 {
		log.Error("ethereum response error, status code: " + strconv.Itoa(resp.StatusCode))
		return "", ""
	}

	body := resp.Body
	defer body.Close()

	response, _ := ioutil.ReadAll(body)
	blockInfo := make(map[string]map[string]string)
	_ = json.Unmarshal(response, &blockInfo)
	if _, ok := blockInfo["error"]; ok {
		log.Error("ethereum response with error")
		log.Error(blockInfo["error"])
		return "", ""
	} else {
		var lastBlockHash string
		lastBlockNumber := blockInfo["result"]["number"][2:]
		if isEven(lastBlockNumber) {
			lastBlockHash = blockInfo["result"]["hash"][2:]
		} else {
			lastBlockHash = blockInfo["result"]["parentHash"][2:]
			lastBlockNumber = subtractOne(lastBlockNumber)
		}
		return lastBlockHash, lastBlockNumber
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

func (e EthereumCollector) getCanonicalFormat(s string) string {
	return s
}

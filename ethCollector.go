package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type EthCollector struct{}

func (e EthCollector) collectEvent() string {
	ethAPI := "http://200.9.100.27/eth"
	jsonStr := []byte(`{"jsonrpc": "2.0", "method": "eth_getBlockByNumber", "id": "1", "params": ["latest", false]}`)
	resp, err := http.Post(ethAPI, "application/json", bytes.NewReader(jsonStr))

	if err != nil {
		// do something
	}

	body := resp.Body
	defer body.Close()

	response, _ := ioutil.ReadAll(body)
	blockInfo := make(map[string](map[string]string))
	_ = json.Unmarshal(response, &blockInfo)
	lastBlockHash := blockInfo["result"]["hash"][2:]
	lastBlockNumber := blockInfo["result"]["number"][2:]

	return lastBlockHash + " " + lastBlockNumber
}

func (e EthCollector) estimateEntropy() int {
	return 0
}

func (e EthCollector) sourceID() int {
	return 1
}

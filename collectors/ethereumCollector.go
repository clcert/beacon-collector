package collectors

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
)

type EthereumCollector struct{}

func (e EthereumCollector) sourceName() string {
	return "ethereum"
}

func (e EthereumCollector) collectEvent(ch chan Event) {
	var status = 0
	sources := []string{"localNode", "infura", "etherscan", "rivet"}
	for _, source := range sources {
		blockHash, blockNumber, valid := getBlock("latest", source)
		if valid {
			if source != "localNode" {
				status = status | FLES_NotDefaultSource
			}
			ch <- Event{blockHash, blockNumber, status}
			return
		}
	}
	log.Error("no ethereum source available")
	ch <- Event{"", "", status | FLES_SourceFail}
}

func getEthToken(source string) string {
	configJsonFile, err := os.Open("collectors/ethereumConfig.json")
	if err != nil {
		fmt.Println(err)
	}
	defer configJsonFile.Close()
	ethTokens := make(map[string]string)
	byteValue, _ := io.ReadAll(configJsonFile)
	json.Unmarshal(byteValue, &ethTokens)

	return ethTokens[source]
}

func isValid(hexNumber string) int64 {
	num, _ := strconv.ParseInt(hexNumber, 16, 64)
	return num % 3
}

func subtractOne(hexNumber string) string {
	num, _ := strconv.ParseInt(hexNumber, 16, 64)
	output := num - 1
	return fmt.Sprintf("%x", output)
}

func getBlock(blockType string, source string) (string, string, bool) {
	var ethAPI string
	switch source {
	case "localNode":
		ethAPI = "https://eth.labs.clcert.cl%s"
	case "infura":
		ethAPI = "https://mainnet.infura.io/v3/%s"
	case "etherscan":
		ethAPI = "https://api.etherscan.io/api?module=proxy&action=eth_getBlockByNumber&tag=" + blockType + "&boolean=false&apikey=%s"
	case "rivet":
		ethAPI = "https://%s.eth.rpc.rivet.cloud/"
	default:
		ethAPI = ""
	}

	jsonStr := []byte(`{"jsonrpc": "2.0", "method": "eth_getBlockByNumber", "id": "1", "params": ["` + blockType + `", false]}`)
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

	response, _ := io.ReadAll(body)
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
		if blockInfo["result"] == nil {
			log.WithFields(log.Fields{
				"ethSource": source,
			}).Error("empty ethereum response")
			return "", "", false
		}
		lastBlockNumber := blockInfo["result"]["number"][2:]
		blockNumberMod := isValid(lastBlockNumber)
		if blockNumberMod == 0 {
			lastBlockHash = blockInfo["result"]["hash"][2:]
			return lastBlockHash, lastBlockNumber, true
		} else if blockNumberMod == 1 {
			parentBlockHash := blockInfo["result"]["parentHash"][2:]
			parentBlockNumber := subtractOne(lastBlockNumber)
			return parentBlockHash, parentBlockNumber, true
		} else {
			greatParentBlockNumber := subtractOne(subtractOne(lastBlockNumber))
			greatParentBlockHash, _, answered := getBlock("0x"+greatParentBlockNumber, source)
			if answered {
				return greatParentBlockHash, greatParentBlockNumber, true
			} else {
				return "", "", false
			}
		}
	}
}

func (e EthereumCollector) estimateEntropy() int {
	return 0
}

func (e EthereumCollector) getCanonicalForm(s string) string {
	return s
}

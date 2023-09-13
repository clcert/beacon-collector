package utils

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"
)

var server = ""

type proveResponse struct {
	Y     string `json:"output"`
	Proof string `json:"proof"`
}

type verifyResponse struct {
	IsValid bool `json:"valid"`
}

func getRandomSeed() []byte {
	seed := make([]byte, 16)
	rand.Read(seed)
	return seed
}

func setServer(newServer string) {
	server = newServer
}

/*
VDFeval function
receives:

	x: input of VDF
	T: number of iterations (squarings)
	ds: discriminant size
	seed: set randomness on discriminant creation

returns:

	(result, proof)
*/
func VDFeval(T, ds int, x, seed []byte) ([]byte, []byte) {

	postBody, _ := json.Marshal(map[string]string{
		"seed":              hex.EncodeToString(seed),
		"input":             hex.EncodeToString(x),
		"iterations":        strconv.Itoa(T),
		"discriminant_size": strconv.Itoa(ds),
	})
	responseBody := bytes.NewBuffer(postBody)

	// Leverage Go's HTTP Post function to make request
	resp, err := http.Post(server+"/eval", "application/json", responseBody)

	// Handle Error
	if err != nil {
		log.Fatalf("An Error Occured %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	decoder := json.NewDecoder(resp.Body)
	var s proveResponse
	err = decoder.Decode(&s)

	if err != nil {
		panic(err)
	}

	y, _ := hex.DecodeString(s.Y)
	p, _ := hex.DecodeString(s.Proof)

	return y, p
}

/*
VERIFY function
receives:

	x: input of VDF
	y: result of VDF
	pi: the proof of VDF result
	T: number of iterations (squarings)
	ds: discriminant size
	seed: set randomness on discriminant creation

returns if verification was correct
*/
func Verify(x, y, pi, seed []byte, T, ds int) bool {

	postBody, _ := json.Marshal(map[string]string{
		"seed":              hex.EncodeToString(seed),
		"input":             hex.EncodeToString(x),
		"output":            hex.EncodeToString(y),
		"proof":             hex.EncodeToString(pi),
		"discriminant_size": strconv.Itoa(ds),
		"iterations":        strconv.Itoa(T),
	})
	responseBody := bytes.NewBuffer(postBody)

	// Leverage Go's HTTP Post function to make request
	resp, err := http.Post(server+"/verify", "application/json", responseBody)

	// Handle Error
	if err != nil {
		log.Fatalf("An Error Occured %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	decoder := json.NewDecoder(resp.Body)
	var s verifyResponse
	err = decoder.Decode(&s)

	if err != nil {
		panic(err)
	}

	return s.IsValid
}

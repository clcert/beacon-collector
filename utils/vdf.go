package utils

import (
	"crypto/rand"

	"github.com/chia-network/go-chia-libs/pkg/vdf"
	log "github.com/sirupsen/logrus"
)

const (
	lambda     uint64 = 1024    // The discriminant size
	iterations uint64 = 2000000 // Also denoted as T
	form_size  int    = 100     // The size of the form
)

func getRandomBytes(nBytes int) []byte {
	seed := make([]byte, nBytes)
	rand.Read(seed)
	return seed
}

func serializeInput(input []byte) []byte {
	if len(input) >= form_size-1 {
		log.Fatalf("Input is too large, must be less than %d bytes", form_size-1)
		log.Exit(1)
	}
	expandedInput := append(make([]byte, form_size-len(input)-1), input...)
	// Prepend the byte 0x08 to the input (VDF requires this byte to be present at the beginning of the input)
	prependByte := byte(0x08)
	return append([]byte{prependByte}, expandedInput...)
}

/*
VDFeval function
receives:

	x: input of VDF
	seed: set randomness on discriminant creation

returns:

	(result, proof)
*/
func VDFeval(x, seed []byte) ([]byte, []byte) {
	serialX := serializeInput(x)
	outVdf := vdf.Prove(seed, serialX, int(lambda), iterations)
	y := outVdf[0:form_size]
	proof := outVdf[form_size:]
	return y, proof
}

/*
VERIFY function
receives:

	x: input of VDF
	y: result of VDF
	pi: the proof of VDF result
	seed: set randomness on discriminant creation

returns if verification was correct
*/
func Verify(x, y, pi, seed []byte) bool {
	// Create discriminant
	discriminant := vdf.CreateDiscriminant(seed, int(lambda))
	// Verify the VDF
	recursion := 0 // We do not use recursion for final output verification
	return vdf.VerifyNWesolowski(discriminant, x, append(y, pi...), iterations, lambda, uint64(recursion))
}

package collectors

import (
	"bufio"
	"encoding/hex"
	"fmt"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/sha3"
	"net/http"
)

type RadioCollector struct{}

func (r RadioCollector) sourceName() string {
	return "radio"
}

const (
	MPEG_2 int = iota
	MPEG_1
)

var SAMPLERATE = map[int]map[byte]int{
	MPEG_1: {
		0x00: 44100,
		0x01: 48000,
		0x02: 32000,
	},
	MPEG_2: {
		0x00: 22050,
		0x01: 12000,
		0x02: 16000,
	},
}
var BITRATE = map[int]map[byte]int{
	MPEG_1: {
		0x01: 32,
		0x02: 40,
		0x03: 48,
		0x04: 56,
		0x05: 64,
		0x06: 80,
		0x07: 96,
		0x08: 112,
		0x09: 128,
		0x0a: 160,
		0x0b: 192,
		0x0c: 224,
		0x0d: 256,
		0x0e: 320,
	},
	MPEG_2: {
		0x01: 8,
		0x02: 16,
		0x03: 24,
		0x04: 32,
		0x05: 40,
		0x06: 48,
		0x07: 56,
		0x08: 64,
		0x09: 80,
		0x0a: 96,
		0x0b: 112,
		0x0c: 128,
		0x0d: 144,
		0x0e: 160,
	},
}

func (r RadioCollector) collectEvent() (string, string) {
	streamURL := "http://200.89.71.21:8000/;"
	resp, _ := http.Get(streamURL)

	if resp == nil {
		log.Error("failed to get radio event")
		return "", ""
	}
	reader := bufio.NewReader(resp.Body)
	var firstFrame []byte
	var audioBytes []byte
	var invalidCounter int
	for frameNumber := 0; frameNumber < 300; frameNumber++ {
		b, _ := reader.ReadByte()
		audioBytes = append(audioBytes, b)
		if frameNumber == 0 {
			firstFrame = append(firstFrame, b)
		}
		if b != 0xff {
			log.Error("invalid sync byte in radio collector")
			return "", ""
		}
		b, _ = reader.ReadByte()
		audioBytes = append(audioBytes, b)
		if frameNumber == 0 {
			firstFrame = append(firstFrame, b)
		}
		if (b & 0xf0) != 0xf0 {
			log.Error("invalid sync byte in radio collector")
			return "", ""
		}
		frameVersion := (b & 0x08) >> 3
		if (b&0x06)>>1 != 1 {
			// Layer is not 3 (0x01)
			log.Error("non layer 3 frame in radio collector")
			return "", ""
		}
		//frameCRC := false
		//if (b & 0x01) == 0x01 {
		//	frameCRC = true
		//}
		b, _ = reader.ReadByte()
		audioBytes = append(audioBytes, b)
		if frameNumber == 0 {
			firstFrame = append(firstFrame, b)
		}
		bitrate := b >> 4
		if bitrate == 0x00 || bitrate == 0x0f {
			// invalid values
			log.Error("invalid bitrate value in radio collector")
			return "", ""
		}
		frameBitRate := BITRATE[int(frameVersion)][bitrate]
		sampleRate := (b & 0x0c) >> 2
		if sampleRate == 0x03 {
			log.Error("invalid samplerate value in radio collector")
			return "", ""
		}
		frameSampleRate := SAMPLERATE[int(frameVersion)][sampleRate]
		padding := (b & 0x02) >> 1
		framePadding := false
		if padding == 1 {
			framePadding = true
		}
		b, _ = reader.ReadByte()
		audioBytes = append(audioBytes, b)
		if frameNumber == 0 {
			firstFrame = append(firstFrame, b)
		}
		frameBodySize := 144000*frameBitRate/frameSampleRate - 4
		if framePadding {
			frameBodySize += 1
		}
		for i := 0; i < frameBodySize; i++ {
			b, _ := reader.ReadByte()
			audioBytes = append(audioBytes, b)
			if frameNumber == 0 {
				firstFrame = append(firstFrame, b)
			}
		}
		// check first frame hash property (first byte equals 0)
		if frameNumber == 0 {
			firstFrameHashedHex := fmt.Sprintf("%x", sha3.Sum512(firstFrame))
			if firstFrameHashedHex[0:2] != "00" {
				invalidCounter += 1
				frameNumber -= 1
				firstFrame = []byte{}
				audioBytes = []byte{}
			}
		}
	}
	// fileMP3, _ := os.Create("audio.mp3")
	// defer fileMP3.Close()
	// fileMP3.Write(audioBytes)
	firstFrameHashedHex := fmt.Sprintf("%x", sha3.Sum512(firstFrame))
	audioHex := hex.EncodeToString(audioBytes)
	return audioHex, firstFrameHashedHex
}

func (r RadioCollector) estimateEntropy() int {
	return 0
}

func (r RadioCollector) getCanonicalForm(s string) string {
	return s
}

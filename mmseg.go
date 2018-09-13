package gommseg

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

var hzRegexp = regexp.MustCompile("^[\u4e00-\u9fa5]$")

var WordMap map[string]*Word

func init() {
	InitWordMap()
}

func InitWordMap() {
	filePath := "./data.txt"

	fp, err := os.Open(filePath)
	if err != nil {
		fmt.Errorf(err.Error())
		os.Exit(0)
	}
	defer fp.Close()

	WordMap = make(map[string]*Word)

	reader := bufio.NewReader(fp)
	for {
		line, _, ok := reader.ReadLine()
		if ok == io.EOF {
			break
		}
		a := bytes.Split(line, []byte("\t"))

		if len(a) >= 2 {
			text := string(a[0])
			freq, e := strconv.Atoi(string(a[1]))
			if e == nil {
				WordMap[text] = NewWord(text, freq)
			}
		}
	}
}

func GetWord(text string) (*Word, bool) {
	word, ok := WordMap[text]
	return word, ok
}

func MatchWords(text string) []*Word {

	var matchWords []*Word
	var matchString string
	//	isFirstPunct := true

	for _, char := range text {
		//		if unicode.IsPunct(char) {

		//		}

		matchString += string(char)
		word, ok := GetWord(matchString)
		if ok {
			matchWords = append(matchWords, word)
		}
	}

	if len(matchWords) == 0 {
		matchWords = append(matchWords, NewWord(matchString, 0))
	}

	return matchWords
}

func Chunks(text string) []*Chunk {
	var chunks []*Chunk
	for _, firstWord := range MatchWords(text) {

		textLength := len(text)
		firstWordLen := len(firstWord.Text)
		if firstWordLen < textLength {
			text1 := string([]byte(text)[firstWordLen:textLength])
			for _, secondWord := range MatchWords(text1) {
				secondWordLen := len(secondWord.Text)
				if firstWordLen+secondWordLen < textLength {
					text2 := string([]byte(text)[firstWordLen+secondWordLen : textLength])
					for _, thirdWord := range MatchWords(text2) {
						chunks = append(chunks, NewChunk([]*Word{firstWord, secondWord, thirdWord}))
					}
				} else {
					chunks = append(chunks, NewChunk([]*Word{firstWord, secondWord}))
				}
			}
		} else {
			chunks = append(chunks, NewChunk([]*Word{firstWord}))
		}
	}

	return chunks
}

func Filter(chunks []*Chunk) *Chunk {
	var maxFilterChunks []*Chunk
	var maxLength int = 0
	for _, chunk := range chunks {
		if chunk.Length() > maxLength {
			maxFilterChunks = []*Chunk{chunk}
			maxLength = chunk.Length()
		} else if chunk.Length() == maxLength {
			maxFilterChunks = append(maxFilterChunks, chunk)
		}
	}

	if len(maxFilterChunks) == 1 {
		return maxFilterChunks[0]
	}

	var averageLengthFilterChunks []*Chunk
	var maxAverageLength float64 = 0
	for _, chunk := range maxFilterChunks {
		if chunk.AverageLength() > maxAverageLength {
			averageLengthFilterChunks = []*Chunk{chunk}
			maxAverageLength = chunk.AverageLength()
		} else if chunk.AverageLength() == maxAverageLength {
			averageLengthFilterChunks = append(averageLengthFilterChunks, chunk)
		}
	}

	if len(averageLengthFilterChunks) == 1 {
		return averageLengthFilterChunks[0]
	}

	var varianceFilterChunks []*Chunk
	var minVariance float64 = 0.0
	for idx, chunk := range averageLengthFilterChunks {
		if idx == 1 {
			varianceFilterChunks = []*Chunk{chunk}
			minVariance = chunk.Variance()
		} else if chunk.Variance() < minVariance {
			varianceFilterChunks = []*Chunk{chunk}
			minVariance = chunk.Variance()
		} else if chunk.Variance() == minVariance {
			varianceFilterChunks = append(varianceFilterChunks, chunk)
		}
	}

	if len(varianceFilterChunks) == 1 {
		return varianceFilterChunks[0]
	}

	var freqFilterChunks []*Chunk
	var maxFreq int = 0
	for _, chunk := range varianceFilterChunks {
		if chunk.Freq() > maxFreq {
			freqFilterChunks = []*Chunk{chunk}
			maxFreq = chunk.Freq()
		} else if chunk.Freq() == maxFreq {
			freqFilterChunks = append(freqFilterChunks, chunk)
		}
	}
	return freqFilterChunks[0]
}

func firstWord(text string) string {
	chunks := Chunks(text)
	chunk := Filter(chunks)
	return chunk.Words[0].Text
}

func Cut(context string) []string {
	var result []string

	sentence := strings.FieldsFunc(context, split)
	for _, text := range sentence {

		textLength := len(text)
		pos := 0
		for pos < textLength {
			str := string([]byte(text)[pos:textLength])
			word := firstWord(str)
			result = append(result, word)
			pos += len(word)
		}
	}
	return result
}

func split(s rune) bool {
	if unicode.IsPunct(s) {
		return true
	}
	if unicode.IsSpace(s) {
		return true
	}
	if s == 1 {
		return true
	}
	if string(s) == "|" {
		return true
	}

	return false
}

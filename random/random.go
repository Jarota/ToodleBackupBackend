package random

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
)

type randomParams struct {
	APIKey     string `json:"apiKey"`
	N          int    `json:"n"`
	Length     int    `json:"length"`
	Characters string `json:"characters"`
}

type randomBody struct {
	JSONrpc string       `json:"jsonrpc"`
	Method  string       `json:"method"`
	Params  randomParams `json:"params"`
	ID      int          `json:"id"`
}

type randomData struct {
	Data           []string `json:"data"`
	CompletionTime string   `json:"completionTime"`
}

type generateStringsResult struct {
	Random        randomData `json:"random"`
	BitsUsed      int        `json:"bitsUsed"`
	BitsLeft      int        `json:"bitsLeft"`
	RequestsLeft  int        `json:"requestsLeft"`
	AdvisoryDelay int        `json:"advisoryDelay"`
}

type randomResponse struct {
	JSONrpc string                `json:"jsonrpc"`
	Result  generateStringsResult `json:"result"`
	ID      int                   `json:"id"`
}

// GetRandomString pings random.org for a true random string
func GetRandomString() (string, error) {

	apiKey := os.Getenv("RANDOMAPI")

	params := &randomParams{
		APIKey:     apiKey,
		N:          1,
		Length:     10,
		Characters: "qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM",
	}

	body := &randomBody{
		JSONrpc: "2.0",
		Method:  "generateStrings",
		Params:  *params,
		ID:      42,
	}

	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(body)

	contentType := "application/json"
	url := "https://api.random.org/json-rpc/2/invoke"
	resp, err := http.Post(url, contentType, buf)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	var randResp randomResponse
	json.Unmarshal(bytes, &randResp)

	rand := randResp.Result.Random.Data[0]
	return rand, nil

}

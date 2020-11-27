package toodledo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/jarota/ToodleBackupBackend/user"
)

type toodleResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	RefreshToken string `json:"refresh_token"`
}

// GetToodledoTokens uses auth code to acquire an access and refresh token
func GetToodledoTokens(code string) (*user.ToodleInfo, error) {

	clientID := "toodlebackup"
	secret := os.Getenv("TOODLEDOSECRET")

	client := &http.Client{}

	apiURL := "https://api.toodledo.com"
	resource := "/3/account/token.php"
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)

	u, _ := url.ParseRequestURI(apiURL)
	u.Path = resource
	urlStr := u.String()

	req, _ := http.NewRequest(http.MethodPost, urlStr, strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	req.SetBasicAuth(clientID, secret)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	fmt.Println(resp.Status)

	bytes, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	fmt.Println(string(bytes))
	var toodleResp toodleResponse
	json.Unmarshal(bytes, &toodleResp)
	// printResponse(&toodleResp)

	return responseToInfo(&toodleResp), nil
}

func printResponse(resp *toodleResponse) {
	fmt.Println("Access Token: " + resp.AccessToken)
	fmt.Println("Token Type: " + resp.TokenType)
	fmt.Println("Scope: " + resp.Scope)
}

func responseToInfo(resp *toodleResponse) *user.ToodleInfo {

	token := resp.AccessToken
	refresh := resp.RefreshToken
	toBackup := strings.Split(resp.Scope, " ")

	return &user.ToodleInfo{
		Token:    token,
		Refresh:  refresh,
		ToBackup: toBackup,
	}
}

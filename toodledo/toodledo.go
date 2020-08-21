package toodledo

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
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
	URL := "https://api.toodledo.com/3/account/token.php"
	v := url.Values{}
	v.Set("grant_type", "authorization_code")
	v.Set("code", code)

	req, err := http.NewRequest("POST", URL, strings.NewReader(v.Encode()))
	req.SetBasicAuth(clientID, secret)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	var toodleResp toodleResponse
	json.Unmarshal(bytes, &toodleResp)

	return responseToInfo(&toodleResp), nil
}

func responseToInfo(resp *toodleResponse) *user.ToodleInfo {

	token := resp.AccessToken
	refresh := resp.RefreshToken
	toBackup := []string{resp.Scope}

	return &user.ToodleInfo{
		Token:    token,
		Refresh:  refresh,
		ToBackup: toBackup,
	}
}

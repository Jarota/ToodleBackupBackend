package dropbox

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/jarota/ToodleBackupBackend/user"
)

type dropboxResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	RefreshToken string `json:"refresh_token"`
	AccountID    string `json:"account_id"`
	UID          string `json:"uid"`
}

// ErrDummyError - Error to throw for testing
var ErrDummyError = errors.New("Error: this is a dummy error")

// GetDropboxTokens gets access and refresh tokens from dropbox
func GetDropboxTokens(code string) (*user.Cloud, error) {

	clientID := "n731o7jng2knpkq"
	clientSecret := os.Getenv("DROPBOXSECRET")

	client := &http.Client{}

	apiURL := "https://api.dropboxapi.com"
	resource := "/oauth2/token"
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", "https://localhost:8080/dropboxredirect")

	u, _ := url.ParseRequestURI(apiURL)
	u.Path = resource
	urlStr := u.String()

	req, _ := http.NewRequest(http.MethodPost, urlStr, strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	req.SetBasicAuth(clientID, clientSecret)

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		bytes, _ := ioutil.ReadAll(resp.Body)
		log.Println(string(bytes))
		return nil, errors.New("Request to connect dropbox failed :(")
	}

	defer resp.Body.Close()

	fmt.Println(resp.Status)

	bytes, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	fmt.Println(string(bytes))
	var dropboxResp dropboxResponse
	json.Unmarshal(bytes, &dropboxResp)

	return responseToCloud(&dropboxResp), nil
}

func responseToCloud(resp *dropboxResponse) *user.Cloud {

	token := resp.RefreshToken

	return &user.Cloud{
		Name:  "Dropbox",
		Token: token,
	}

}

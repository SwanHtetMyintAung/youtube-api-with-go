package Youtube

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"youtube-service/Helper"
)

type GoogleClient struct {
	Installed struct {
		ClientID     string   `json:"client_id"`
		RedirectURIs []string `json:"redirect_uris"`
		Scopes       []string `json:"scopes"`
		ResponseType string
		AccessType   string
		ClientSecret string `json:"client_secret"`
	} `json:"installed"`
}
type OAuthToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

const BASE_URL = "https://accounts.google.com/o/oauth2/v2/auth" //nlm go, fuck you. i'll use JS conventions here
const OAUTH2_TOKEN_URL = "https://oauth2.googleapis.com/token"

// Init : will return a token to do api calls
func Init() string {
	//check if the tokens already exists
	var token OAuthToken
	var client GoogleClient
	_, err := LoadTokenFromFile("token.json", &token)

	if err != nil {
		println("Could not load token.json and proceeding to normal procedures.")
		err = CreateTokenFromConsentScreen("token.json", "client.json", &token, &client)
		if err != nil {
			println(err.Error())
			os.Exit(1)
		}
	}
	valid, err := CheckValidityOfToken(token)
	if valid == false {
		println("the token isn't valid, refreshing the token.")
		//refresh the token here
		err = RefreshAccessToken(&token, client)
		if err != nil {
			println(err.Error())
		} else {
			//if the program wasn't able to refresh the token, make the user sign in again
			println("Could not load token.json and proceeding to normal procedures.")
			err = CreateTokenFromConsentScreen("token.json", "client.json", &token, &client)
			if err != nil {
				println(err.Error())
				os.Exit(1)
			}
		}

	} else {
		println("the token is valid")
	}

	return token.AccessToken
}

func CreateTokenFromConsentScreen(tokenFilePath string, clientFilePath string, token *OAuthToken, client *GoogleClient) error {
	url, err := MakeConsentScreenUrl(clientFilePath, client)
	if err != nil {
		return err
	}
	println(url)

	var code string
	print("Please enter a code here: ")
	fmt.Scan(&code)

	data, err := GetTokensFromCode(*client, code)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &token)
	if err != nil {
		return err
	}
	err = os.WriteFile(tokenFilePath, data, 0644)
	if err != nil {
		fmt.Println("Could not write token.json")
	}
	return nil
}

func GetPlaylistItems(url string, token string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return []byte(""), err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return []byte(""), err
	}

	if resp.StatusCode != 200 {
		return []byte(""), errors.New("Could not get playlist items.")
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte(""), err
	}

	return data, nil
}

// LoadTokenFromFile : true if the file is found and tokens are there
func LoadTokenFromFile(filepath string, token *OAuthToken) (bool, error) {
	tokenBytes, err := os.ReadFile(filepath)
	if err != nil {
		return false, err
	}

	err = json.Unmarshal(tokenBytes, &token)
	if err != nil {
		return false, err
	}
	if token.RefreshToken != "" && token.AccessToken != "" {
		return true, nil
	}
	return false, errors.New("invalid token")
}

func MakeConsentScreenUrl(filepath string, client *GoogleClient) (string, error) {
	credByte, err := os.ReadFile(filepath)

	if err != nil {
		return "", err
	}
	jsonErr := json.Unmarshal(credByte, &client)

	if jsonErr != nil {
		return "", jsonErr
	}
	//this part is kinda clunky
	client.Installed.ResponseType = "code"
	client.Installed.Scopes = []string{"https://www.googleapis.com/auth/youtube"}
	client.Installed.AccessType = "offline"

	urlTOReturn := BASE_URL + "?client_id=" + client.Installed.ClientID + "&redirect_uri=" + client.Installed.RedirectURIs[0] + "&" + client.Installed.ResponseType + "&scope=" + client.Installed.Scopes[0] + "&access_type=" + client.Installed.AccessType + "&response_type=" + client.Installed.ResponseType
	return urlTOReturn, nil
}

// GetTokensFromCode : will return the "body" of the http response from Youtube
func GetTokensFromCode(client GoogleClient, code string) ([]byte, error) {
	tokenJson := map[string]string{
		"code":          code,
		"client_id":     client.Installed.ClientID,
		"client_secret": client.Installed.ClientSecret,
		"redirect_uri":  client.Installed.RedirectURIs[0],
		"grant_type":    "authorization_code",
	}
	tokenByte, err := json.Marshal(tokenJson)
	if err != nil {
		return []byte(""), err
	}
	resp, err := http.Post(OAUTH2_TOKEN_URL, "application/json", bytes.NewBuffer(tokenByte))
	if err != nil {
		return []byte(""), err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return data, nil
}

func CheckValidityOfToken(token OAuthToken) (bool, error) {
	resp, err := http.Get("https://oauth2.googleapis.com/tokeninfo?access_token=" + token.AccessToken)
	if err != nil {
		return false, err
	}
	return resp.StatusCode == 200, nil
}
func RefreshAccessToken(token *OAuthToken, client GoogleClient) error {
	tokenJson := map[string]string{
		"client_id":     client.Installed.ClientID,
		"client_secret": client.Installed.ClientSecret,
		"refresh_token": token.RefreshToken,
		"grant_type":    "refresh_token",
	}
	//change the map into []bytes
	tokenByte, err := json.Marshal(tokenJson)
	if err != nil {
		return err
	}
	resp, err := http.Post(OAUTH2_TOKEN_URL, "application/json", bytes.NewBuffer(tokenByte))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	json.Unmarshal(data, &token)
	return nil
}

// DeletePlaylistItem : as the name suggest
func DeletePlaylistItems(access_token string, playlist_id string) error { //not the video id
	req, err := http.NewRequest("DELETE",
		"https://www.googleapis.com/youtube/v3/playlistItems?id="+playlist_id,
		nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+access_token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func ClearPlaylist(token string) error {
	var playlistURL string
	print("Playlist url: ")
	fmt.Scan(&playlistURL)
	playlistId := Helper.ExtractIdForPlaylist(playlistURL)
	urlToPlaylist := "https://www.googleapis.com/youtube/v3/playlistItems" + "?part=snippet" + "&playlistId=" + playlistId + "&maxResults=30"

	resBody, err := GetPlaylistItems(urlToPlaylist, token)

	if err != nil {
		return err
	}
	//temp
	os.WriteFile("playlistItems.json", resBody, 0644)
	//use the resBody ([]byte) and cast it or use the data directly
	var test map[string]interface{}
	err = json.Unmarshal(resBody, &test)
	videos := test["items"].([]interface{})
	for _, video := range videos {
		videoMap := video.(map[string]interface{})
		playlist_id := videoMap["id"].(string)

		err = DeletePlaylistItems(token, playlist_id)
		if err != nil {
			return err
		}
	}

	println("It seems like it was successful.")
	return nil
}

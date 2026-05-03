package Download

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"youtube-service/Helper"
	"youtube-service/Youtube"
)

// Run is the entry point for the download package
func Run() {

	PrintYtdlpMenu()
	var num int8
	for num != 3 {
		//user choice
		_, err := fmt.Scanf("%d", &num)
		if err != nil {
			fmt.Println(err)
			return
		}

		switch num {
		case 1: //single video
			reader := bufio.NewReader(os.Stdin)
			print("Enter a video url :")
			videoURL, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Input error:", err)
				return
			}
			videoURL = strings.TrimSpace(videoURL)

			DownloadSingleVideo(videoURL)

		case 2: //playlist
			reader := bufio.NewReader(os.Stdin)
			print("Enter a playlist url :")
			playlistURL, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Input error:", err)
				return
			}
			playlistURL = strings.TrimSpace(playlistURL)

			DownloadPlaylist(playlistURL)
		case 3:
			return
		default:
			println("You Choice isn't valid.")
		}
	}
}

// --- dependency checks ---
func checkDeps() {
	if !Helper.CommandExists("python3") && !Helper.CommandExists("python") {
		fmt.Println("Python is not installed or not in PATH")
		os.Exit(1)
	}

	if !Helper.CommandExists("yt-dlp") {
		fmt.Println("yt-dlp is not installed or not in PATH")
		os.Exit(1)
	}
}

func PrintYtdlpMenu() {
	println("Enter 1 to download a single video")
	println("Enter 2 to download the whole playlist")
	println("Enter 3 to go back")

	print("Your choice :")
}

func DownloadSingleVideo(url string) {
	//Example: yt-dlp <url>
	//yt-dlp -x --audio-format mp3 <URL>
	cmd := exec.Command("yt-dlp" + "-x --audio-format mp3" + "-o \"%(title)s.%(ext)s\"" + url)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return
}
func DownloadPlaylist(playlistUrl string) {
	playlistId := Helper.ExtractIdForPlaylist(playlistUrl)

	if playlistId == "" {
		println("The Url didn't contain the id")
		return
	}
	playlistUrl = "https://www.googleapis.com/youtube/v3/playlistItems" + "?part=snippet" + "&playlistId=" + playlistId + "&maxResults=30"

	var token Youtube.OAuthToken
	_, err := Youtube.LoadTokenFromFile("token.json", &token)
	if err != nil {
		println(err.Error())
		println("You need to login")
		return
	}
	itemBytes, err := Youtube.GetPlaylistItems(playlistUrl, token.AccessToken)

	if err != nil {
		println("Could not get playlist items.")
		os.Exit(1)
	}
	//use the resBody ([]byte) and cast it or use the data directly
	var songs map[string]interface{}
	err = json.Unmarshal(itemBytes, &songs) //error here
	videos := songs["items"].([]interface{})
	for _, video := range videos {
		videoMap := video.(map[string]interface{})
		contentDetails := videoMap["snippet"].(map[string]interface{})
		//resourceId
		resourceId := contentDetails["resourceId"].(map[string]interface{})

		videoID := resourceId["videoId"].(string)

		videoUrl := "https://www.youtube.com/watch?v=" + videoID
		println(videoUrl)
		DownloadSingleVideo(videoUrl)
	}

}

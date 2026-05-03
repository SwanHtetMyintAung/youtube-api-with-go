package Helper

import (
	"net/url"
	"os/exec"
)

func ExtractIdForPlaylist(playlistUrl string) string {
	httpUrl, err := url.Parse(playlistUrl)
	if err != nil {
		println(err.Error())
		return ""
	}
	return httpUrl.Query().Get("list")
}
func CommandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"ssl-manager/config"
	"strings"

	"golang.org/x/text/message"
)

func SendNotification(config.NotificationWebhook, certRenewError error){

callWebhook("https://gotify.bekucera.uk/message?token=A34AtJnsFC3aNZR", map[string]string{"title": "ssl-manager", "message": "%message%"})
}

func callWebhook(url string, postData map[string]string) error {
	jsonBody, err := json.Marshal(postData)
	if err != nil {
		return err
	}
	bodyReader := bytes.NewReader(jsonBody)

	resp, err := http.Post(url, "application/json", bodyReader)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("sending notification fialed, status code %d, response: %s", resp.StatusCode, resp.Body)
	}

	return nil
}

func replaceMapVariables(mapToReplace *map[string]string){
	for k := range(mapToReplace){
		mapToReplace[k] = strings.ReplaceAll(*mapToReplace[k], "%message%")
	}	
}

package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"ssl-manager/config"
	"strings"
)

func SendCertRenewNotifications(webhooks []config.NotificationWebhook, renewedCerts []string, certRenewError error) error {
	for _, v := range webhooks {
		if (certRenewError == nil && v.NotifySuccess) || (certRenewError != nil && v.NotifyFail) {
			err := SendCertRenewNotification(&v, renewedCerts, certRenewError)
			if err != nil {
				return err
			}
		}
	}
	return nil

}

func SendCertRenewNotification(webhook *config.NotificationWebhook, renewedCerts []string, certRenewError error) error {
	err := callWebhook(webhook.Url, replaceMapVariables(webhook.PostData, renewedCerts, certRenewError))
	if err != nil {
		return err
	}
	return nil
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
		return fmt.Errorf("sending http request fialed, status code %d, response: %s", resp.StatusCode, resp.Body)
	}

	return nil
}

func replaceMapVariables(mapToReplace map[string]string, renewedCerts []string, certRenewError error) map[string]string {
	mapToReplaceCopy := make(map[string]string)
	for k := range mapToReplace {
		var message string
		var status string
		if certRenewError == nil {
			message = fmt.Sprintf("Certificates were renewed successfully (%s)", strings.Join(renewedCerts, ", "))
			status = "success"
		} else {
			message = certRenewError.Error()
			status = "failed"
		}
		mapToReplaceCopy[k] = mapToReplace[k]
		mapToReplaceCopy[k] = strings.ReplaceAll(mapToReplaceCopy[k], "%message%", message)
		mapToReplaceCopy[k] = strings.ReplaceAll(mapToReplaceCopy[k], "%status%", status)

	}

	return mapToReplaceCopy
}

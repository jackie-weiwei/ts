package ts

import (
	"encoding/json"
	"fmt"

	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/token"
)

var (
	apnsClient  *apns2.Client
	isInit      bool
	appBundleId string
)

func ApnsInit(certificate []byte, keyID string, teamID string, appBundle string) {
	appBundleId = appBundle
	authKey, err := token.AuthKeyFromBytes(certificate)
	if err != nil {
		fmt.Println(err)
	}

	t := &token.Token{AuthKey: authKey, KeyID: keyID, TeamID: teamID}

	apnsClient = apns2.NewTokenClient(t).Production().Development()
	isInit = true
}

func ApnsPush(token string, msg string) error {
	if !isInit {
		return fmt.Errorf("apns not init, call ApnsInit() first")
	}
	payload := make(map[string]interface{})
	aps := make(map[string]interface{})

	aps["alert"] = msg
	aps["sound"] = "defalut"
	aps["badge"] = 1
	payload["aps"] = aps

	notification := &apns2.Notification{}
	notification.DeviceToken = token
	notification.Topic = appBundleId

	jsonPayload, _ := json.Marshal(payload)

	notification.Payload = jsonPayload

	res, err := apnsClient.Push(notification)

	if res.StatusCode != 200 {
		return fmt.Errorf("apns push error, status code: %d, %s", res.StatusCode, res.Reason)
	}
	return err
}

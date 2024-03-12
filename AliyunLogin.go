package ts

import (
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dypnsapi20170525 "github.com/alibabacloud-go/dypnsapi-20170525/v2/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
)

func createAliyunLoginClient(accessKeyId *string, accessKeySecret *string) (_result *dypnsapi20170525.Client, _err error) {
	config := &openapi.Config{
		AccessKeyId:     accessKeyId,
		AccessKeySecret: accessKeySecret,
	}

	config.Endpoint = tea.String("dypnsapi.aliyuncs.com")

	_result, _err = dypnsapi20170525.NewClient(config)
	return _result, _err
}

func AliyunGetMobile(token string, accessKeyId string, accessKeySecret string) (string, error) {
	client, err := createAliyunLoginClient(tea.String(accessKeyId), tea.String(accessKeySecret))

	if err != nil {
		return "", err
	}

	getMobileRequest := &dypnsapi20170525.GetMobileRequest{
		AccessToken: tea.String(token),
	}
	runtime := &util.RuntimeOptions{}

	response, _err := client.GetMobileWithOptions(getMobileRequest, runtime)

	if _err == nil {

		bok := response.Body.Code

		if *bok == "OK" {
			strMobile := response.Body.GetMobileResultDTO.Mobile

			return *strMobile, nil

		} else {
			return "", nil
		}

	} else {
		return "", _err
	}

}

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	arms20190808 "github.com/alibabacloud-go/arms-20190808/v11/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	credential "github.com/aliyun/credentials-go/credentials"
	"github.com/joho/godotenv"
)

func CreateClient() (_result *arms20190808.Client, _err error) {
	// 工程代码建议使用更安全的无AK方式，凭据配置方式请参见：https://help.aliyun.com/document_detail/378661.html。
	cfg := credential.Config{}
	cfg.SetType("access_key")
	cfg.SetAccessKeyId(os.Getenv("BUGLENS_ALIBABA_ACCESS_KEY_ID"))
	cfg.SetAccessKeySecret(os.Getenv("BUGLENS_ALIBABA_ACCESS_KEY_SECRET"))
	credential, _err := credential.NewCredential(&cfg)
	if _err != nil {
		return _result, _err
	}

	config := &openapi.Config{
		Credential: credential,
	}
	// Endpoint 请参考 https://api.aliyun.com/product/ARMS
	config.Endpoint = tea.String("arms.cn-hangzhou.aliyuncs.com")
	_result = &arms20190808.Client{}
	_result, _err = arms20190808.NewClient(config)
	return _result, _err
}

func _main(args []*string) (_err error) {
	client, _err := CreateClient()
	if _err != nil {
		return _err
	}

	getRumExceptionStackRequest := &arms20190808.GetRumExceptionStackRequest{
		Pid:                   tea.String("a7q597fa88@3ecf5e6c7b91ec7"),
		ExceptionStack:        tea.String("245,16085,20"),
		ExceptionBinaryImages: tea.String("{\"version\":\"1.0.0\",\"fileName\":\"index-DXuEeTYs.js.map\",\"uuid\":\"21be7c88-e56c-4014-b08f-f5b2095dcef1\"}"),
		RegionId:              tea.String("cn-hangzhou"),
		SourcemapType:         tea.String("js"),
	}
	runtime := &util.RuntimeOptions{}
	tryErr := func() (_e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
			}
		}()
		resp, _err := client.GetRumExceptionStackWithOptions(getRumExceptionStackRequest, runtime)
		if _err != nil {
			return _err
		}

		fmt.Printf("[LOG] %v\n", resp)

		return nil
	}()

	if tryErr != nil {
		var error = &tea.SDKError{}
		if _t, ok := tryErr.(*tea.SDKError); ok {
			error = _t
		} else {
			error.Message = tea.String(tryErr.Error())
		}
		// 此处仅做打印展示，请谨慎对待异常处理，在工程项目中切勿直接忽略异常。
		// 错误 message
		fmt.Println(tea.StringValue(error.Message))
		// 诊断地址
		var data interface{}
		d := json.NewDecoder(strings.NewReader(tea.StringValue(error.Data)))
		d.Decode(&data)
		if m, ok := data.(map[string]interface{}); ok {
			recommend, _ := m["Recommend"]
			fmt.Println(recommend)
		}
	}
	return _err
}

func main() {
	godotenv.Load(".env")
	err := _main(tea.StringSlice(os.Args[1:]))
	if err != nil {
		panic(err)
	}
}

package sms

import (
	"encoding/json"
	"errors"
	"fmt"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dypnsapi20170525 "github.com/alibabacloud-go/dypnsapi-20170525/v3/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"go.uber.org/zap"
	"strings"
	response "tuyou/internal/model/dto/response/web/auth"
	"tuyou/pkg/config"
	appErrors "tuyou/pkg/errors"
)

// Description:
//
// 使用凭据初始化账号Client
//
// @return Client
//
// @throws Exception

// CreateClient
// 推荐使用阿里云 Credentials 自动读取凭证（环境变量、RAM角色等），
// 不建议在代码中直接写 AccessKey。
//
// 返回值：
//   - *dypnsapi20170525.Client：SDK客户端
//   - error：初始化失败时返回错误
func CreateClient(sms config.SMSConfig) (_result *dypnsapi20170525.Client, _err error) {
	// 工程代码建议使用更安全的无AK方式，凭据配置方式请参见：https://help.aliyun.com/document_detail/378661.html。SDK 会自动从环境变量、RAM角色等方式读取凭证
	// 创建凭据对象
	// 此处使用的是AK方式，AK存在了配置文件里
	//credentialsConfig := new(credential.Config).
	//	SetType("access_key").
	//	SetAccessKeyId(sms.AccessKeyId).
	//	SetAccessKeySecret(sms.AccessKeySecret)
	//client, _err := credential.NewCredential(credentialsConfig)
	//if _err != nil {
	//	return _result, _err
	//}
	// 创建 OpenAPI 配置
	cfg := &openapi.Config{
		AccessKeyId:     tea.String(sms.AccessKeyId),
		AccessKeySecret: tea.String(sms.AccessKeySecret),
	}
	// Endpoint 请参考 https://api.aliyun.com/product/Dypnsapi
	// 设置接口 Endpoint
	// 官方文档：https://api.aliyun.com/product/Dypnsapi
	cfg.Endpoint = tea.String("dypnsapi.aliyuncs.com")
	_result = &dypnsapi20170525.Client{}
	// 创建客户端
	_result, _err = dypnsapi20170525.NewClient(cfg)
	return _result, _err
}

// Send 发送验证码
func Send(phone string) (*response.MessageResponse, error) {
	var sms = config.Get().SMS

	// 创建SDK客户端
	client, _err := CreateClient(sms)
	if _err != nil {
		return nil, _err
	}
	// 创建发送验证码请求对象
	// 这里需要填写手机号、验证码模板等参数
	sendSmsVerifyCodeRequest := &dypnsapi20170525.SendSmsVerifyCodeRequest{
		SchemeName:       tea.String(sms.SchemeName),
		CountryCode:      tea.String(sms.CountryCode),
		PhoneNumber:      tea.String(phone),
		SignName:         tea.String(sms.SignName),
		TemplateCode:     tea.String(sms.TemplateCode),
		TemplateParam:    tea.String(sms.TemplateParam),
		CodeLength:       tea.Int64(sms.CodeLength),
		ValidTime:        tea.Int64(sms.ValidTime),
		DuplicatePolicy:  tea.Int64(sms.DuplicatePolicy),
		Interval:         tea.Int64(sms.Interval),
		CodeType:         tea.Int64(sms.CodeType),
		ReturnVerifyCode: tea.Bool(sms.ReturnVerifyCode),
		AutoRetry:        tea.Int64(sms.AutoRetry),
	}
	// SDK运行配置
	// 可配置超时时间、代理等
	runtime := &util.RuntimeOptions{}
	// 调用接口
	tryErr := func() (_e error) {
		// SDK异常恢复
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
			}
		}()
		// 发送验证码
		resp, _err1 := client.SendSmsVerifyCodeWithOptions(sendSmsVerifyCodeRequest, runtime)
		if _err1 != nil {
			return _err1
		}
		if resp != nil && resp.Body != nil && !tea.BoolValue(resp.Body.Success) {
			code := tea.StringValue(resp.Body.Code)
			msg := tea.StringValue(resp.Body.Message)
			if code == "biz.FREQUENCY" {
				return appErrors.ErrSmsSendFrequently
			}
			return appErrors.New(appErrors.CodeSmsSendFrequently, "短信发送失败："+msg)
		}

		return nil
	}()
	// 如果调用失败
	if tryErr != nil {
		var sdkErr = &tea.SDKError{}
		if errors.As(tryErr, &sdkErr) {
			// SDK 错误，尝试解析诊断信息
			zap.L().Error("send sms verify code error" + sdkErr.Error())
			var data interface{}
			d :=
				json.NewDecoder(strings.NewReader(tea.StringValue(sdkErr.Data)))
			if err2 := d.Decode(&data); err2 == nil {
				if m, ok := data.(map[string]interface{}); ok {
					recommend, _ := m["Recommend"]
					fmt.Println(recommend)
				}
			}
		}
		// 无论是不是 SDK 错误，都要返回 tryErr
		return nil, tryErr
	}

	return &response.MessageResponse{Message: "验证码已发送"}, nil
}

// Receive 接收验证码
func Receive(phone string, code string) (_err error) {
	var sms = config.Get().SMS
	client, _err := CreateClient(sms)
	if _err != nil {
		return _err
	}

	checkSmsVerifyCodeRequest := &dypnsapi20170525.CheckSmsVerifyCodeRequest{
		SchemeName:     tea.String(sms.SchemeName),
		CountryCode:    tea.String(sms.CountryCode),
		PhoneNumber:    tea.String(phone),
		CaseAuthPolicy: tea.Int64(sms.CaseAuthPolicy),
		VerifyCode:     tea.String(code),
	}
	runtime := &util.RuntimeOptions{}
	tryErr := func() (_e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
			}
		}()
		resp, err := client.CheckSmsVerifyCodeWithOptions(checkSmsVerifyCodeRequest, runtime)
		if err != nil {
			return err
		}

		if resp != nil && resp.Body != nil && !tea.BoolValue(resp.Body.Success) {
			return fmt.Errorf("验证码校验失败：%s - %s", tea.StringValue(resp.Body.Code), tea.StringValue(resp.Body.Message))
		}

		fmt.Printf("[LOG] %v\n", resp)

		return nil
	}()

	if tryErr != nil {
		var err1 = &tea.SDKError{}
		if errors.As(tryErr, &err1) {
			// err 已经是 SDKError
		} else {
			err1.Message = tea.String(tryErr.Error())
		}
		// 此处仅做打印展示，请谨慎对待异常处理，在工程项目中切勿直接忽略异常。
		// 错误 message
		fmt.Println(tea.StringValue(err1.Message))
		// 诊断地址
		var data interface{}
		d := json.NewDecoder(strings.NewReader(tea.StringValue(err1.Data)))
		err := d.Decode(&data)
		if err != nil {
			zap.L().Error("receive sms verify code error" + err.Error())
			return err
		}
		if m, ok := data.(map[string]interface{}); ok {
			recommend, _ := m["Recommend"]
			fmt.Println(recommend)
		}
	}
	return tryErr
}

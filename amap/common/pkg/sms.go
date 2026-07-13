package pkg

import (
	"common/config"
	"fmt"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dypnsapi "github.com/alibabacloud-go/dypnsapi-20170525/v3/client"
	"github.com/alibabacloud-go/tea/tea"
)

// Sms 发送短信验证码（阿里云号码认证·短信认证），返回平台生成的验证码。
func Sms(tel string) (string, error) {
	data := config.DataConfig.AliSms
	cfg := &openapi.Config{
		AccessKeyId:     tea.String(data.AccessKeyId),
		AccessKeySecret: tea.String(data.AccessKeySecret),
		Endpoint:        tea.String("dypnsapi.aliyuncs.com"),
	}
	client, err := dypnsapi.NewClient(cfg)
	if err != nil {
		return "", err
	}

	req := &dypnsapi.SendSmsVerifyCodeRequest{
		PhoneNumber:      tea.String(tel),
		SignName:         tea.String(data.SignName),
		TemplateCode:     tea.String(data.TemplateCode),
		TemplateParam:    tea.String(`{"code":"##code##","min":"5"}`),
		CodeType:         tea.Int64(1),
		CodeLength:       tea.Int64(4),
		ReturnVerifyCode: tea.Bool(true),
		ValidTime:        tea.Int64(180),
	}

	resp, err := client.SendSmsVerifyCode(req)
	if err != nil {
		return "", err
	}

	if resp.Body == nil || tea.StringValue(resp.Body.Code) != "OK" {
		code := ""
		msg := ""
		if resp.Body != nil {
			code = tea.StringValue(resp.Body.Code)
			msg = tea.StringValue(resp.Body.Message)
		}
		if code != "" && msg != "" {
			return "", fmt.Errorf("短信发送失败: %s (%s)", msg, code)
		}
		return "", fmt.Errorf("短信发送失败")
	}

	if resp.Body.Model == nil || resp.Body.Model.VerifyCode == nil {
		return "", fmt.Errorf("短信发送失败: 未返回验证码")
	}
	return tea.StringValue(resp.Body.Model.VerifyCode), nil
}

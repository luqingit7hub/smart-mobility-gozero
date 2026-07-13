package pkg

import (
	"common/config"
	"context"
	"fmt"
	"net/url"

	"github.com/smartwalle/alipay/v3"
)

func AlipayPass(form url.Values) bool {
	data := config.DataConfig.Alipay
	client, err := alipay.New(data.AppId, data.PrivateKey, true)
	if err != nil {
		fmt.Println(err)
		return false
	}
	if err := client.LoadAliPayPublicKey(data.AliPubKey); err != nil {
		fmt.Println(err)
		return false
	}
	fmt.Println("网络中解析到的:", form)
	if err := client.VerifySign(context.Background(), form); err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

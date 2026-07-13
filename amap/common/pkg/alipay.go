package pkg

import (
	"common/config"
	"fmt"

	"github.com/smartwalle/alipay/v3"
)

func AliPay(tradeNo string, price float64) string {
	conf := config.DataConfig.Alipay
	client, err := alipay.New(conf.AppId, conf.PrivateKey, false)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	p := alipay.TradeWapPay{}
	p.NotifyURL = conf.NotifyUrl
	p.ReturnURL = "http://baidu.com"
	p.Subject = "用户余额充值"
	p.OutTradeNo = tradeNo
	p.TotalAmount = fmt.Sprintf("%.2f", price)
	p.ProductCode = "QUICK_WAP_WAY"

	url, err := client.TradeWapPay(p)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	payURL := url.String()
	fmt.Println(payURL)
	return payURL
}

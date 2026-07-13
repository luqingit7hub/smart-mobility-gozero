package pkg

import (
	"common/config"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	gourl "net/url"
	"strings"
	"time"

	"github.com/hashicorp/go-uuid"
)

func calcAuthorization(secretId string, secretKey string) (auth string, datetime string, err error) {
	timeLocation, _ := time.LoadLocation("Etc/GMT")
	datetime = time.Now().In(timeLocation).Format("Mon, 02 Jan 2006 15:04:05 GMT")
	signStr := fmt.Sprintf("x-date: %s", datetime)

	// hmac-sha1
	mac := hmac.New(sha1.New, []byte(secretKey))
	mac.Write([]byte(signStr))
	sign := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	auth = fmt.Sprintf("{\"id\":\"%s\", \"x-date\":\"%s\", \"signature\":\"%s\"}",
		secretId, datetime, sign)

	return auth, datetime, nil
}

func urlencode(params map[string]string) string {
	var p = gourl.Values{}
	for k, v := range params {
		p.Add(k, v)
	}
	return p.Encode()
}

func RealName(cardNo, realName string) (T2, error) {
	data := config.DataConfig.SmRz
	// 云市场分配的密钥Id
	secretId := data.SecretID
	// 云市场分配的密钥Key
	secretKey := data.SecretKey
	// 签名
	auth, _, _ := calcAuthorization(secretId, secretKey)

	// 请求方法
	method := "POST"
	reqID, err := uuid.GenerateUUID()
	if err != nil {
		return T2{}, err
	}
	// 请求头
	headers := map[string]string{"Authorization": auth, "request-id": reqID}

	// 查询参数
	queryParams := make(map[string]string)

	// body参数
	bodyParams := make(map[string]string)
	bodyParams["cardNo"] = strings.TrimSpace(cardNo)
	bodyParams["realName"] = strings.TrimSpace(realName)
	bodyParamStr := urlencode(bodyParams)
	// url参数拼接
	url := "https://ap-beijing.cloudmarket-apigw.com/service-18c38npd/idcard/VerifyIdcardv2"

	if len(queryParams) > 0 {
		url = fmt.Sprintf("%s?%s", url, urlencode(queryParams))
	}

	bodyMethods := map[string]bool{"POST": true, "PUT": true, "PATCH": true}
	var body io.Reader = nil
	if bodyMethods[method] {
		body = strings.NewReader(bodyParamStr)
		headers["Content-Type"] = "application/x-www-form-urlencoded"
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return T2{}, err
	}
	for k, v := range headers {
		request.Header.Set(k, v)
	}
	response, err := client.Do(request)
	if err != nil {
		return T2{}, err
	}
	defer response.Body.Close()

	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return T2{}, err
	}
	//fmt.Println(string(bodyBytes))
	var t T2
	json.Unmarshal(bodyBytes, &t)
	return t, err
}

type T2 struct {
	ErrorCode int    `json:"error_code"`
	Reason    string `json:"reason"`
	Result    struct {
		Realname    string `json:"realname"`
		Idcard      string `json:"idcard"`
		Isok        bool   `json:"isok"`
		IdCardInfor struct {
			Province string `json:"province"`
			City     string `json:"city"`
			District string `json:"district"`
			Area     string `json:"area"`
			Sex      string `json:"sex"`
			Birthday string `json:"birthday"`
		} `json:"IdCardInfor"`
	} `json:"result"`
}
type T3 struct {
	ErrorCode int    `json:"error_code"`
	Reason    string `json:"reason"`
	Result    struct {
		Realname    string      `json:"realname"`
		Idcard      string      `json:"idcard"`
		Isok        bool        `json:"isok"`
		IdCardInfor interface{} `json:"IdCardInfor"`
	} `json:"result"`
}

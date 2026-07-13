package pkg

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	BAIDUAK              = "JtXxKkAVkDuqO6S4WeqeLlqjy2ywgTCL"
	BAIDUCOORDINATESHOST = "https://api.map.baidu.com"
	BAIDUCOORDINATESURL  = "/geocoding/v3/"
	BAIDUPATHPLANHOST    = "https://api.map.baidu.com"
	BAIDUPATHPLANURL     = "/directionlite/v1/driving"
)

type Coordinates struct {
	Status int    `json:"status"`
	Result Result `json:"result"`
}

type Result struct {
	Location      Location `json:"location"`
	Precise       int      `json:"precise"`
	Confidence    int      `json:"confidence"`
	Comprehension int      `json:"comprehension"`
	Level         string   `json:"level"`
}

type Location struct {
	Lng float64 `json:"lng"`
	Lat float64 `json:"lat"`
}

func GetCoordinates(address string) (Location, error) { //位置获取经纬度
	if address == "" {
		return Location{}, errors.New("地址不能为空")
	}
	ak := BAIDUAK
	host := BAIDUCOORDINATESHOST
	uri := BAIDUCOORDINATESURL

	params := url.Values{
		"address": []string{address},
		"output":  []string{"json"},
		"ak":      []string{ak},
	}

	request, err := url.Parse(host + uri + "?" + params.Encode())
	if err != nil {
		return Location{}, fmt.Errorf("host error %w", err)
	}

	resp, err := http.Get(request.String())
	if err != nil {
		return Location{}, fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Location{}, fmt.Errorf("response error: %w", err)
	}

	var content Coordinates
	if err := json.Unmarshal(body, &content); err != nil {
		return Location{}, errors.New("JSON解析失败")
	}
	if content.Status != 0 {
		return Location{}, geocodeStatusError(content.Status)
	}

	return content.Result.Location, nil
}

func geocodeStatusError(status int) error {
	switch status {
	case 1:
		return errors.New("服务器内部错误")
	case 2:
		return errors.New("请求参数非法")
	case 3:
		return errors.New("权限校验失败")
	case 4:
		return errors.New("配额校验失败")
	case 5:
		return errors.New("ak不存在或者非法")
	case 101:
		return errors.New("AK参数不存在")
	case 102:
		return errors.New("不通过白名单或者安全码不对")
	case 240:
		return errors.New("APP 服务被禁用")
	default:
		return errors.New("地理编码失败")
	}
}

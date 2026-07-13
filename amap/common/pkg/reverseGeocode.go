package pkg

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type reverseGeocodeAPIResult struct {
	Status int `json:"status"`
	Result struct {
		FormattedAddress string `json:"formatted_address"`
	} `json:"result"`
}

// ReverseGeocode 根据百度 bd09ll 坐标获取可读地址
func ReverseGeocode(lng, lat float64) (string, error) {
	if lng == 0 && lat == 0 {
		return "", errors.New("经纬度不能为空")
	}
	params := url.Values{
		"ak":        []string{BAIDUAK},
		"output":    []string{"json"},
		"coordtype": []string{"bd09ll"},
		"location":  []string{fmt.Sprintf("%f,%f", lat, lng)},
	}
	request, err := url.Parse(BAIDUCOORDINATESHOST + "/reverse_geocoding/v3/?" + params.Encode())
	if err != nil {
		return "", fmt.Errorf("url parse error: %w", err)
	}
	resp, err := http.Get(request.String())
	if err != nil {
		return "", fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body error: %w", err)
	}
	var result reverseGeocodeAPIResult
	if err := json.Unmarshal(body, &result); err != nil {
		return "", errors.New("逆地理编码JSON解析失败")
	}
	if result.Status != 0 {
		return "", errors.New("百度逆地理编码API返回错误")
	}
	addr := result.Result.FormattedAddress
	if addr == "" {
		return "", errors.New("未解析到地址")
	}
	return addr, nil
}

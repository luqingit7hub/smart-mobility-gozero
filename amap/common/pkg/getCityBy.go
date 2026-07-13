package pkg

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// reverseGeocodingResult 百度逆地理编码响应
type reverseGeocodingResult struct {
	Status int `json:"status"`
	Result struct {
		AddressComponent struct {
			Province string `json:"province"`
			City     string `json:"city"`
			Adcode   string `json:"adcode"` // 行政区划代码，与优惠券 city_code 比对
		} `json:"addressComponent"`
	} `json:"result"`
}

// GetCityByLocation 根据用户经纬度获取行政区划代码与城市名（优惠券地区校验用）
func GetCityByLocation(lng, lat float64) (adcode string, cityName string, err error) {
	if lng == 0 && lat == 0 {
		return "", "", errors.New("经纬度不能为空")
	}
	params := url.Values{
		"ak":        []string{BAIDUAK},
		"output":    []string{"json"},
		"coordtype": []string{"bd09ll"},
		"location":  []string{fmt.Sprintf("%f,%f", lat, lng)}, // 百度 API：纬度在前
	}
	request, err := url.Parse(BAIDUCOORDINATESHOST + "/reverse_geocoding/v3/?" + params.Encode())
	if err != nil {
		return "", "", fmt.Errorf("url parse error: %w", err)
	}
	resp, err := http.Get(request.String())
	if err != nil {
		return "", "", fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("read body error: %w", err)
	}
	var result reverseGeocodingResult
	if err := json.Unmarshal(body, &result); err != nil {
		return "", "", errors.New("逆地理编码JSON解析失败")
	}
	if result.Status != 0 {
		return "", "", errors.New("百度逆地理编码API返回错误")
	}
	adcode = result.Result.AddressComponent.Adcode
	cityName = result.Result.AddressComponent.City
	if cityName == "" {
		cityName = result.Result.AddressComponent.Province // 直辖市 city 可能为空
	}
	return adcode, cityName, nil
}

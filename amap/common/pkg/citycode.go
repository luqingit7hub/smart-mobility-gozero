package pkg

import "errors"

// CityLevelAdcode 将区县 adcode 归一到市级（321302 → 321300）
func CityLevelAdcode(adcode string) string {
	if len(adcode) < 4 {
		return adcode
	}
	return adcode[:4] + "00"
}

// MatchCouponCityCode 判断用户位置是否可用该城市券（区县相等或同市）
func MatchCouponCityCode(couponCityCode, userAdcode string) bool {
	if couponCityCode == "" || userAdcode == "" {
		return false
	}
	if couponCityCode == userAdcode {
		return true
	}
	return CityLevelAdcode(couponCityCode) == CityLevelAdcode(userAdcode)
}

// AdcodeFromAddress 地址文本 → 百度地理编码 → 逆地理 adcode（与发券/用券一致）
func AdcodeFromAddress(address string) (adcode string, cityName string, err error) {
	if address == "" {
		return "", "", errors.New("地址不能为空")
	}
	loc, err := GetCoordinates(address)
	if err != nil {
		return "", "", err
	}
	return GetCityByLocation(loc.Lng, loc.Lat)
}

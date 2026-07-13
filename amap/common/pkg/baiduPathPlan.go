package pkg

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type RouteResponse struct {
	Status  int            `json:"status"`
	Message string         `json:"message"`
	Result  PathPlanResult `json:"result"`
}

type PathPlanResult struct {
	Routes []BaiduRoute `json:"routes"`
}

type BaiduRoute struct {
	Distance         int         `json:"distance"`
	Duration         int         `json:"duration"`
	TrafficCondition int         `json:"traffic_condition"`
	Toll             float64     `json:"toll"`
	Steps            []RouteStep `json:"steps"`
}

type RouteStep struct {
	Path string `json:"path"`
}

type SimpleRoute struct {
	Distance         int
	Duration         int
	TrafficCondition int
	Toll             float64
	RoutePoints      []Location
}

type PathPlanReq struct {
	OriginLng      float64
	OriginLat      float64
	DestinationLng float64
	DestinationLat float64
}

func parsePathPoints(path string) []Location {
	if path == "" {
		return nil
	}
	var pts []Location
	for _, seg := range strings.Split(path, ";") {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}
		parts := strings.Split(seg, ",")
		if len(parts) != 2 {
			continue
		}
		lng, err1 := strconv.ParseFloat(parts[0], 64)
		lat, err2 := strconv.ParseFloat(parts[1], 64)
		if err1 != nil || err2 != nil {
			continue
		}
		pts = append(pts, Location{Lng: lng, Lat: lat})
	}
	return pts
}

func collectRoutePoints(steps []RouteStep) []Location {
	var all []Location
	for _, step := range steps {
		all = append(all, parsePathPoints(step.Path)...)
	}
	return all
}

func GetPathPlan(form PathPlanReq) ([]SimpleRoute, error) {
	ak := BAIDUAK
	host := BAIDUPATHPLANHOST
	uri := BAIDUPATHPLANURL
	origin := fmt.Sprintf("%f,%f", form.OriginLat, form.OriginLng)
	destination := fmt.Sprintf("%f,%f", form.DestinationLat, form.DestinationLng)

	params := url.Values{
		"origin":      []string{origin},
		"destination": []string{destination},
		"ak":          []string{ak},
		"steps_info":  []string{"1"},
	}

	request, err := url.Parse(host + uri + "?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("host error %w", err)
	}

	resp, err := http.Get(request.String())
	if err != nil {
		return nil, fmt.Errorf("request error %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("request error %w", err)
	}

	var ditu RouteResponse
	if err = json.Unmarshal(body, &ditu); err != nil {
		return nil, errors.New("JSON解析失败")
	}

	if ditu.Status != 0 {
		msg := ""
		switch ditu.Status {
		case 1:
			msg = "服务内部错误"
		case 2:
			msg = "参数无效"
		case 7:
			msg = "无返回结果"
		default:
			msg = "请求失败"
		}
		return nil, errors.New(msg)
	}

	var list []SimpleRoute
	for _, v := range ditu.Result.Routes {
		distanceKm := float64(v.Distance) / 1000
		durationMin := v.Duration / 60
		originalToll := v.Toll

		const (
			baseFare   = 10.0
			baseKm     = 3.0
			pricePerKm = 2.5
		)

		orderPrice := baseFare
		if distanceKm > baseKm {
			orderPrice += (distanceKm - baseKm) * pricePerKm
		}
		orderPrice += originalToll

		list = append(list, SimpleRoute{
			Distance:         int(distanceKm),
			Duration:         durationMin,
			TrafficCondition: v.TrafficCondition,
			Toll:             orderPrice,
			RoutePoints:      collectRoutePoints(v.Steps),
		})
	}
	return list, nil
}

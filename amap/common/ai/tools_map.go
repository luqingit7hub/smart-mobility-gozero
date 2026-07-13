package ai

import (
	"context"
	"encoding/json"

	"common/pkg"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type geocodeInput struct {
	Address string `json:"address" jsonschema_description:"中文地址，例如：北京市海淀区中关村"`
}

func geocode(_ context.Context, input *geocodeInput) (string, error) {
	loc, err := pkg.GetCoordinates(input.Address)
	if err != nil {
		return "", err
	}
	b, _ := json.Marshal(loc)
	return string(b), nil
}

type pathPlanInput struct {
	OriginLng      float64 `json:"origin_lng" jsonschema_description:"起点经度"`
	OriginLat      float64 `json:"origin_lat" jsonschema_description:"起点纬度"`
	DestinationLng float64 `json:"destination_lng" jsonschema_description:"终点经度"`
	DestinationLat float64 `json:"destination_lat" jsonschema_description:"终点纬度"`
}

func pathPlan(_ context.Context, input *pathPlanInput) (string, error) {
	routes, err := pkg.GetPathPlan(pkg.PathPlanReq{
		OriginLng:      input.OriginLng,
		OriginLat:      input.OriginLat,
		DestinationLng: input.DestinationLng,
		DestinationLat: input.DestinationLat,
	})
	if err != nil {
		return "", err
	}
	b, _ := json.Marshal(routes)
	return string(b), nil
}

// NewMapTools 返回地图相关 Tool，用户和司机都能用。
func NewMapTools() ([]tool.BaseTool, error) {
	geocodeTool, err := utils.InferTool("geocode", "根据中文地址获取经纬度", geocode)
	if err != nil {
		return nil, err
	}
	pathPlanTool, err := utils.InferTool("path_plan", "根据起终点经纬度规划驾车路线", pathPlan)
	if err != nil {
		return nil, err
	}
	return []tool.BaseTool{geocodeTool, pathPlanTool}, nil
}

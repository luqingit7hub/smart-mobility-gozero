package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// BizBackend 业务数据查询接口，由 rpcMap 的 MapChatLogic 实现。
type BizBackend interface {
	GetMyBalance(ctx context.Context) (rawJSON string, err error)
	ListMyOrders(ctx context.Context, page, pageSize int32) (rawJSON string, err error)
	ListCoupons(ctx context.Context, couponType int64) (rawJSON string, err error)
	PreviewJourney(ctx context.Context, start, end string) (rawJSON string, err error)
	GrabList(ctx context.Context, radiusM int64, limit int32) (rawJSON string, err error)
}

// BizChatWithFallback 业务助手：先意图直答，未命中再走 Agent，Agent 客套时再次直答兜底。
func BizChatWithFallback(ctx context.Context, question string, uid int64, role int, backend BizBackend) (string, error) {
	ctx = WithSession(ctx, uid, role)

	if ans, ok, err := tryDirectBizAnswer(ctx, backend, question, role); ok {
		if err != nil {
			return "", err
		}
		return ans, nil
	}

	bizTools, err := newBizTools(backend)
	if err != nil {
		return "", err
	}

	answer, err := BizChat(ctx, question, uid, role, bizTools)
	if err == nil && !IsAckOnlyAnswer(answer) {
		return answer, nil
	}

	if ans, ok, fbErr := tryDirectBizAnswer(ctx, backend, question, role); ok {
		if fbErr != nil {
			return "", fbErr
		}
		return ans, nil
	}
	if err != nil {
		return "", err
	}
	return answer, nil
}

type bizIntent int

const (
	intentNone bizIntent = iota
	intentJourney
	intentBalance
	intentOrders
	intentCoupons
	intentGrabList
)

func detectBizIntent(question string, role int) bizIntent {
	q := strings.TrimSpace(question)
	if q == "" {
		return intentNone
	}
	if start, end, ok := parseJourneyRoute(q); ok && start != "" && end != "" {
		return intentJourney
	}
	if role == 2 && isGrabListQuestion(q) {
		return intentGrabList
	}
	if role == 1 && containsAny(q, "优惠券", "代金券", "我的券", "有哪些券") {
		return intentCoupons
	}
	if isOrderQuestion(q) {
		return intentOrders
	}
	if isBalanceQuestion(q) {
		return intentBalance
	}
	return intentNone
}

func isBalanceQuestion(q string) bool {
	if strings.Contains(q, "从") && strings.Contains(q, "到") {
		return false
	}
	if containsAny(q, "余额", "账户余额", "钱包", "我的余额", "还有多少钱", "还有多少", "有多少钱", "还有钱", "剩多少", "还剩") {
		return true
	}
	// 「多少钱」仅在明确问自身资金时命中，避免误判车费询价。
	if strings.Contains(q, "多少钱") {
		return containsAny(q, "我", "余额", "账户", "钱包", "剩")
	}
	return false
}

func isOrderQuestion(q string) bool {
	if strings.Contains(q, "从") && strings.Contains(q, "到") {
		return false
	}
	return containsAny(q, "订单", "接单记录", "我的单", "历史订单", "查看订单", "我的订单")
}

func isGrabListQuestion(q string) bool {
	if strings.Contains(q, "从") && strings.Contains(q, "到") {
		return false
	}
	return containsAny(q, "抢单", "有什么单", "可抢", "能抢", "附近有什么单", "附近可抢", "附近订单")
}

// parseJourneyRoute 解析「从起点到终点」类询价，如：从北京西站到天安门多少钱
func parseJourneyRoute(q string) (start, end string, ok bool) {
	idx := strings.Index(q, "从")
	if idx < 0 {
		return "", "", false
	}
	rest := q[idx+len("从"):]
	toIdx := strings.Index(rest, "到")
	if toIdx <= 0 {
		return "", "", false
	}
	start = strings.TrimSpace(rest[:toIdx])
	end = trimJourneyTail(strings.TrimSpace(rest[toIdx+len("到"):]))
	if start == "" || end == "" {
		return "", "", false
	}
	return start, end, true
}

func trimJourneyTail(s string) string {
	for {
		changed := false
		for _, suffix := range []string{
			"多少钱", "多少费用", "多少路费", "多少车费", "打车多少钱",
			"要花多少", "费用多少", "价格多少", "路费", "车费", "打车费",
			"要多久", "多远", "怎么走", "怎么样",
		} {
			if strings.HasSuffix(s, suffix) {
				s = strings.TrimSpace(strings.TrimSuffix(s, suffix))
				changed = true
			}
		}
		s = strings.Trim(s, "？?！!。 ")
		if !changed {
			break
		}
	}
	return strings.TrimSpace(s)
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

func tryDirectBizAnswer(ctx context.Context, backend BizBackend, question string, role int) (string, bool, error) {
	switch detectBizIntent(question, role) {
	case intentJourney:
		if role != 1 {
			return "", false, nil
		}
		start, end, ok := parseJourneyRoute(question)
		if !ok {
			return "", false, nil
		}
		raw, err := backend.PreviewJourney(ctx, start, end)
		if err != nil {
			return "", true, err
		}
		return formatJourneyAnswer(start, end, raw), true, nil
	case intentBalance:
		raw, err := backend.GetMyBalance(ctx)
		if err != nil {
			return "", true, err
		}
		return formatBalanceAnswer(raw), true, nil
	case intentOrders:
		raw, err := backend.ListMyOrders(ctx, 1, 5)
		if err != nil {
			return "", true, err
		}
		return formatOrdersAnswer(raw), true, nil
	case intentCoupons:
		if role != 1 {
			return "", false, nil
		}
		raw, err := backend.ListCoupons(ctx, 0)
		if err != nil {
			return "", true, err
		}
		return formatCouponsAnswer(raw), true, nil
	case intentGrabList:
		if role != 2 {
			return "", false, nil
		}
		raw, err := backend.GrabList(ctx, 0, 0)
		if err != nil {
			return "", true, err
		}
		return formatGrabListAnswer(raw), true, nil
	default:
		return "", false, nil
	}
}

func formatBalanceAnswer(raw string) string {
	var v struct {
		Balance float64 `json:"balance"`
	}
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return "暂时无法解析余额数据，请稍后再试。"
	}
	return fmt.Sprintf("您当前账户余额为 ¥%.2f。", v.Balance)
}

func formatJourneyAnswer(start, end, raw string) string {
	var v struct {
		Price    float32 `json:"price"`
		Distance int64   `json:"distance"`
		Duration int64   `json:"duration"`
	}
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return "暂时无法解析估价数据，请稍后再试。"
	}
	return fmt.Sprintf(
		"从 %s 到 %s 预估车费约 ¥%.2f，全程约 %d 公里，预计 %d 分钟。",
		start, end, v.Price, v.Distance, v.Duration,
	)
}

func formatOrdersAnswer(raw string) string {
	var resp struct {
		List []struct {
			StartAddress string  `json:"start_address"`
			EndAddress   string  `json:"end_address"`
			Status       int32   `json:"status"`
			StatusName   string  `json:"status_name"`
			Price        float64 `json:"price"`
		} `json:"list"`
		Total int64 `json:"total"`
	}
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return "暂时无法解析订单数据，请稍后再试。"
	}
	if len(resp.List) == 0 {
		return "您目前没有订单记录。"
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("您共有 %d 笔订单，最近 %d 笔如下：\n", resp.Total, len(resp.List)))
	for i, o := range resp.List {
		status := o.StatusName
		if status == "" {
			status = fmt.Sprintf("状态%d", o.Status)
		}
		b.WriteString(fmt.Sprintf("%d. %s → %s，%s，¥%.2f\n", i+1, o.StartAddress, o.EndAddress, status, o.Price))
	}
	return strings.TrimSpace(b.String())
}

func formatCouponsAnswer(raw string) string {
	var resp struct {
		List []struct {
			TypeName  string  `json:"type_name"`
			MoneyQuan float64 `json:"money_quan"`
			Discount  float64 `json:"discount"`
			OutTime   string  `json:"out_time"`
		} `json:"list"`
		Count int32 `json:"count"`
	}
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return "暂时无法解析优惠券数据，请稍后再试。"
	}
	if len(resp.List) == 0 {
		return "您目前没有可用优惠券。"
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("您共有 %d 张优惠券：\n", resp.Count))
	for i, c := range resp.List {
		name := c.TypeName
		if name == "" {
			name = "优惠券"
		}
		if c.MoneyQuan > 0 {
			b.WriteString(fmt.Sprintf("%d. %s 满减 ¥%.0f，有效期至 %s\n", i+1, name, c.MoneyQuan, c.OutTime))
		} else if c.Discount > 0 {
			b.WriteString(fmt.Sprintf("%d. %s 折扣 %.1f 折，有效期至 %s\n", i+1, name, c.Discount*10, c.OutTime))
		} else {
			b.WriteString(fmt.Sprintf("%d. %s，有效期至 %s\n", i+1, name, c.OutTime))
		}
	}
	return strings.TrimSpace(b.String())
}

func formatGrabListAnswer(raw string) string {
	var resp struct {
		Orders []struct {
			StartAddress     string  `json:"startAddress"`
			EndAddress       string  `json:"endAddress"`
			Price            float64 `json:"price"`
			DistanceToDriver float64 `json:"distanceToDriver"`
		} `json:"orders"`
	}
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return "暂时无法解析抢单列表，请稍后再试。"
	}
	if len(resp.Orders) == 0 {
		return "附近暂无可抢订单，请稍后再看。"
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("附近共有 %d 单可抢：\n", len(resp.Orders)))
	for i, o := range resp.Orders {
		b.WriteString(fmt.Sprintf("%d. %s → %s，¥%.2f，距您 %.0fm\n", i+1, o.StartAddress, o.EndAddress, o.Price, o.DistanceToDriver))
	}
	return strings.TrimSpace(b.String())
}

type emptyInput struct{}

type listOrderInput struct {
	Page     int32 `json:"page" jsonschema_description:"页码，默认1"`
	PageSize int32 `json:"page_size" jsonschema_description:"每页条数，默认10，最大20"`
}

type couponInput struct {
	Type int64 `json:"type" jsonschema_description:"优惠券类型，0表示全部"`
}

type journeyInput struct {
	Start string `json:"start" jsonschema_description:"起点地址"`
	End   string `json:"end" jsonschema_description:"终点地址"`
}

type grabListInput struct {
	RadiusM int64 `json:"radius_m" jsonschema_description:"搜索半径(米)，默认5000"`
	Limit   int32 `json:"limit" jsonschema_description:"返回单数，默认10"`
}

func newBizTools(backend BizBackend) ([]tool.BaseTool, error) {
	var tools []tool.BaseTool

	myBalanceTool, err := utils.InferTool("get_my_balance", "查询当前登录用户或司机的账户余额", func(ctx context.Context, _ *emptyInput) (string, error) {
		return backend.GetMyBalance(ctx)
	})
	if err != nil {
		return nil, err
	}

	myOrdersTool, err := utils.InferTool("list_my_orders", "查询当前登录用户或司机的最近订单", func(ctx context.Context, input *listOrderInput) (string, error) {
		page := input.Page
		if page <= 0 {
			page = 1
		}
		pageSize := input.PageSize
		if pageSize <= 0 {
			pageSize = 10
		}
		if pageSize > 20 {
			pageSize = 20
		}
		return backend.ListMyOrders(ctx, page, pageSize)
	})
	if err != nil {
		return nil, err
	}

	couponTool, err := utils.InferTool("list_coupons", "查询当前用户优惠券（仅乘客）", func(ctx context.Context, input *couponInput) (string, error) {
		return backend.ListCoupons(ctx, input.Type)
	})
	if err != nil {
		return nil, err
	}

	journeyTool, err := utils.InferTool("preview_journey", "估算从起点到终点的打车费用和距离（仅乘客）", func(ctx context.Context, input *journeyInput) (string, error) {
		return backend.PreviewJourney(ctx, input.Start, input.End)
	})
	if err != nil {
		return nil, err
	}

	grabListTool, err := utils.InferTool("grab_list", "查询司机附近可抢订单列表（仅司机）", func(ctx context.Context, input *grabListInput) (string, error) {
		radius := input.RadiusM
		if radius <= 0 {
			radius = 5000
		}
		limit := input.Limit
		if limit <= 0 {
			limit = 10
		}
		return backend.GrabList(ctx, radius, limit)
	})
	if err != nil {
		return nil, err
	}

	tools = append(tools, myBalanceTool, myOrdersTool, couponTool, journeyTool, grabListTool)
	return tools, nil
}

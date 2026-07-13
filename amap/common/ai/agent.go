package ai

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
)

var (
	chatModel model.ToolCallingChatModel
	chatAgent adk.Agent
	mapTools  []tool.BaseTool
)

// Init 服务启动时调用一次（在 rpcMap 的 main 里调）。
func Init(ctx context.Context) error {
	var err error
	chatModel, err = newChatModel(ctx)
	if err != nil {
		return err
	}

	mapTools, err = NewMapTools()
	if err != nil {
		return err
	}

	chatAgent, err = adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "ChatAssistant",
		Description: "通用聊天助手",
		Instruction: "你是友好助手，用简洁中文回答。不要假装能查询余额、订单等系统数据。",
		Model:       chatModel,
	})
	return err
}

// Chat type=1：纯聊天，不调 Tool。
func Chat(ctx context.Context, question string) (string, error) {
	return runAgent(ctx, chatAgent, question)
}

// BizChat type=2：业务助手，合并地图 Tool + extraTools（由 rpcMap 传入）。
func BizChat(ctx context.Context, question string, uid int64, role int, extraTools []tool.BaseTool) (string, error) {
	if chatModel == nil {
		return "", fmt.Errorf("AI Agent 未初始化，请先调用 ai.Init")
	}
	ctx = withSession(ctx, uid, role)

	tools := make([]tool.BaseTool, 0, len(mapTools)+len(extraTools))
	tools = append(tools, mapTools...)
	tools = append(tools, extraTools...)

	bizAgent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "BizAssistant",
		Description: "网约车业务助手",
		Instruction: `你是网约车项目智能助手。
规则：
1. 查余额、订单、优惠券、抢单、路线时，必须先调用工具，禁止编造。
2. 当前登录身份已由系统注入，不要向用户索要 uid。
3. 问余额时调用 get_my_balance；问订单时调用 list_my_orders。
4. 乘客还可调用 list_coupons、preview_journey；司机还可调用 grab_list。
5. 地图工具 geocode、path_plan 用户和司机都能用。
6. 用简洁中文回答,用户问的问题必须要给出回答,无法实现可以回答不会`,
		Model: chatModel,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: tools,
			},
		},
		MaxIterations: 12,
	})
	if err != nil {
		return "", err
	}
	answer, err := runAgent(ctx, bizAgent, question)
	if err != nil {
		return "", err
	}
	if IsAckOnlyAnswer(answer) {
		return "", fmt.Errorf("模型未调用工具")
	}
	return answer, nil
}

// IsAckOnlyAnswer 判断是否为「只答应不办事」的客套回复。
func IsAckOnlyAnswer(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	if strings.Contains(s, "¥") || strings.Contains(s, "余额为") ||
		(strings.Contains(s, "元") && strings.ContainsAny(s, "0123456789")) {
		return false
	}
	ackWords := []string{"我来", "帮您", "帮你", "正在", "查询", "查一下", "稍等", "马上"}
	matched := 0
	for _, w := range ackWords {
		if strings.Contains(s, w) {
			matched++
		}
	}
	return matched >= 2 && len([]rune(s)) < 40
}

func runAgent(ctx context.Context, agent adk.Agent, question string) (string, error) {
	if agent == nil {
		return "", fmt.Errorf("AI Agent 未初始化")
	}

	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           agent,
		EnableStreaming: false,
	})

	iter := runner.Query(ctx, question)
	var answer string

	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if event.Err != nil {
			return "", event.Err
		}
		if event.Output == nil || event.Output.MessageOutput == nil {
			continue
		}
		mv := event.Output.MessageOutput
		if mv.IsStreaming {
			stream := mv.MessageStream
			if stream == nil {
				continue
			}
			for {
				msg, err := stream.Recv()
				if err == io.EOF {
					break
				}
				if err != nil {
					return "", err
				}
				if msg != nil && msg.Content != "" {
					answer += msg.Content
				}
			}
		} else if mv.Message != nil && mv.Message.Content != "" {
			answer = mv.Message.Content
		}
	}

	if answer == "" {
		return "", fmt.Errorf("模型未返回内容")
	}
	return answer, nil
}

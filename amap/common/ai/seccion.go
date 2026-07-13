package ai

import "context"

type ctxKey int

const sessionKey ctxKey = 1

// Session 保存当前登录用户/司机信息，Tool 里用来查「我的」数据。
type Session struct {
	UID  int64
	Role int // 1=乘客 2=司机
}

// WithSession 注入当前登录身份，供 BizChat 与 rpcMap 直调业务查询使用。
func WithSession(ctx context.Context, uid int64, role int) context.Context {
	return context.WithValue(ctx, sessionKey, &Session{UID: uid, Role: role})
}

func withSession(ctx context.Context, uid int64, role int) context.Context {
	return WithSession(ctx, uid, role)
}

// Tool 里取出 uid / role，决定调乘客还是司机接口
func sessionFrom(ctx context.Context) (*Session, bool) {
	s, ok := ctx.Value(sessionKey).(*Session)
	return s, ok
}

// SessionFrom 供 rpcMap 业务 Tool 读取当前登录身份。
func SessionFrom(ctx context.Context) (*Session, bool) {
	return sessionFrom(ctx)
}

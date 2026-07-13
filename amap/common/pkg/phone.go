package pkg

import "strings"

// MatchPhoneTail 校验手机号后四位（仅数字，长度必须为 4）
func MatchPhoneTail(phone, tail string) bool {
	phone = strings.TrimSpace(phone)
	tail = strings.TrimSpace(tail)
	if len(tail) != 4 || len(phone) < 4 {
		return false
	}
	for _, c := range tail {
		if c < '0' || c > '9' {
			return false
		}
	}
	return phone[len(phone)-4:] == tail
}

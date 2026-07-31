package service

import (
	"sync"
	"time"
)

// tokenBlacklist 登出 token 黑名单（单机内存方案，重启即失效）
var tokenBlacklist sync.Map // token → exp unix

// BlacklistToken 将 token 加入黑名单直至过期
func BlacklistToken(token string, exp time.Time) {
	tokenBlacklist.Store(token, exp.Unix())
}

// IsTokenBlacklisted 检查 token 是否在黑名单且未过期；过期项惰性清理
func IsTokenBlacklisted(token string) bool {
	v, ok := tokenBlacklist.Load(token)
	if !ok {
		return false
	}
	if time.Now().Unix() > v.(int64) {
		tokenBlacklist.Delete(token)
		return false
	}
	return true
}

// CleanExpiredTokens 清理已过期黑名单项（登录时调用）
func CleanExpiredTokens() {
	now := time.Now().Unix()
	tokenBlacklist.Range(func(k, v interface{}) bool {
		if now > v.(int64) {
			tokenBlacklist.Delete(k)
		}
		return true
	})
}

package router

import (
	"testing"

	"probig/server/internal/config"
	"probig/server/internal/middleware"
	"probig/server/internal/model"
)

// TestRouterPermissionKeysWithinDefinitions 权限键一致性防错网：
// 路由（RequirePermission）使用的全部权限键必须定义于 model.ModuleActions。
// 新增业务模块时若路由键拼错/漏定义，本测试立即失败。
func TestRouterPermissionKeysWithinDefinitions(t *testing.T) {
	// SetupRouter 依赖 config.AppConfig（CORS 中间件构造时读取），先初始化默认配置
	config.LoadConfig()

	// SetupRouter() 构造全部 RequirePermission 闭包，权限键登记完毕
	_ = SetupRouter()

	defined := make(map[string]bool)
	for _, mod := range model.ModuleActions {
		for _, action := range mod.Actions {
			defined[mod.Module+"."+action] = true
		}
	}
	for _, key := range middleware.RegisteredPermissionKeys() {
		if !defined[key] {
			t.Errorf("路由权限键 %s 未在 model.ModuleActions 中定义", key)
		}
	}
}

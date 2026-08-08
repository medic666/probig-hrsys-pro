package model

// PermissionActionNames 权限动作显示名（动作类型唯一来源：查看/编辑(增删改)/核算/导出）。
// 权限体系底座：任何模块 × 动作的组合均由此定义驱动（种子与迁移共用）。
var PermissionActionNames = map[string]string{
	"read":      "查看",
	"write":     "编辑",
	"calculate": "核算",
	"export":    "导出",
}

// ModuleActions 模块 × 动作定义（模块唯一来源）：种子与迁移共用，
// 新增模块/动作只改这一处。
var ModuleActions = []struct {
	Module  string
	Name    string
	Actions []string
}{
	{"home", "首页", []string{"read"}},
	{"person", "人员管理", []string{"read", "write", "export"}},
	{"company", "公司管理", []string{"read", "write", "export"}},
	{"position_event", "职务事件", []string{"read", "write", "export"}},
	{"attendance", "考勤管理", []string{"read", "write", "calculate", "export"}},
	{"annual_leave", "年假管理", []string{"read", "write", "calculate", "export"}},
	{"salary", "工资管理", []string{"read", "write", "calculate", "export"}},
	{"file", "文件管理", []string{"read", "write", "export"}},
	{"audit", "审计日志", []string{"read", "export"}},
	{"user", "用户管理", []string{"read", "write"}},
	{"role", "角色管理", []string{"read", "write"}},
	{"system_config", "系统配置", []string{"read", "write"}},
}

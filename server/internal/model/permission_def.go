package model

// PermissionActionNames 权限动作显示名（动作类型唯一来源：查看/编辑(增删改)/核算/导出）。
// 权限体系底座：任何模块 × 动作的组合均由此定义驱动（种子与迁移共用）。
var PermissionActionNames = map[string]string{
	"read":      "查看",
	"write":     "编辑",
	"calculate": "核算",
	"export":    "导出",
}

// ModuleActions 模块 × 动作定义（模块唯一来源，与前端菜单叶子同构）：
// 种子与迁移共用，新增模块/动作只改这一处。
var ModuleActions = []struct {
	Module  string
	Name    string
	Actions []string
}{
	{"person", "人员管理", []string{"read", "write", "export"}},
	{"company", "公司管理", []string{"read", "write", "export"}},
	{"position_event", "职务事件", []string{"read", "write", "export"}},
	{"attendance_event", "考勤事件", []string{"read", "write", "export"}},
	{"attendance_daily", "日记工时", []string{"read", "export"}},
	{"attendance_monthly", "月度考勤核算", []string{"read", "calculate", "export"}},
	{"annual_leave_event", "年假事件", []string{"read", "write", "export"}},
	{"leave_in_lieu", "调休事件", []string{"read"}},
	{"annual_leave_carryover", "年假配发结转", []string{"read", "calculate"}},
	{"salary_event", "工资事件", []string{"read", "write", "export"}},
	{"salary_summary", "月度工资汇总", []string{"read", "calculate", "export"}},
	{"user", "用户管理", []string{"read", "write"}},
	{"role", "角色管理", []string{"read", "write"}},
	{"system_config", "系统配置", []string{"read", "write"}},
	{"file", "文件管理", []string{"read", "write"}},
	{"audit", "审计日志", []string{"read", "export"}},
}

// ReferenceEndpoints 结构授权点清单：跨模块「人员×时间」主轴底座数据。
// 仅需认证即可访问（路由挂 middleware.StructureOnly），不参与模块×动作角色授权，
// 不进入角色权限分配 UI。安全边界：只暴露选项/卡片级参考信息，业务明细永远在业务授权点内。
var ReferenceEndpoints = []struct {
	Method string
	Path   string
	Desc   string
}{
	{"GET", "/api/persons/all", "人员选项（人员选择组件数据源）"},
	{"GET", "/api/persons/cards", "人员卡片（卡片视图底座）"},
	{"GET", "/api/companies/all", "公司选项（公司选择组件数据源）"},
}

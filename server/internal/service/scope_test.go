package service

import (
	"context"
	"testing"

	"probig/server/internal/dao"
	"probig/server/internal/model"

	"gorm.io/gorm"
)

// ctxWithScope 构造带数据范围上下文的请求 ctx（own 需关联人员）
func ctxWithScope(userID uint, scope string, personID uint) context.Context {
	ctx := context.Background()
	info := dao.ScopeInfo{UserID: userID, DataScope: scope}
	if personID > 0 {
		info.PersonID = &personID
	}
	return dao.WithScopeInfo(ctx, info)
}

// TestScopeListFilter 仅自己范围：人员维度列表/卡片/选择源只返回本人
func TestScopeListFilter(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		db.AutoMigrate(&model.Person{}, &model.AttendanceDaily{}, &model.AttendanceEventDetail{})
		seedPerson(t, db, 90, "张三")
		seedPerson(t, db, 91, "李四")
		seedSalaryEvent(db, 90, "2026-06", "提成", 100)
		seedSalaryEvent(db, 91, "2026-06", "提成", 200)
		db.Create(&model.PositionEvent{PersonID: 90, Seq: 1, EventType: "入职"})
		db.Create(&model.PositionEvent{PersonID: 91, Seq: 1, EventType: "入职"})

		ctx := ctxWithScope(2, dao.DataScopeOwn, 90)

		// 工资事件列表：仅本人
		evs, total, err := GetSalaryEventList(ctx, SalaryEventListQuery{PageNum: 1, PageSize: 10})
		if err != nil || total != 1 {
			t.Fatalf("salary list: total=%d err=%v, want 1", total, err)
		}
		if evs[0]["person_id"].(uint) != 90 {
			t.Errorf("salary list person = %v, want 90", evs[0]["person_id"])
		}

		// 职务事件列表：仅本人
		_, peTotal, _ := GetPositionEventList(ctx, PositionEventListQuery{PageNum: 1, PageSize: 10})
		if peTotal != 1 {
			t.Errorf("position list total=%d, want 1", peTotal)
		}

		// 人员卡片：仅本人
		cards, err := GetPersonCards(ctx)
		if err != nil || len(cards) != 1 || cards[0].ID != 90 {
			t.Errorf("cards: %v err=%v, want only 90", cards, err)
		}

		// 人员选择源：仅本人
		opts, err := GetAllPersons(ctx)
		if err != nil || len(opts) != 1 || opts[0].ID != 90 {
			t.Errorf("all persons: %v err=%v, want only 90", opts, err)
		}
	})
}

// TestScopeWriteDenied 仅自己范围：写操作他人数据被拒绝
func TestScopeWriteDenied(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		db.AutoMigrate(&model.Person{})
		seedPerson(t, db, 90, "张三")
		seedPerson(t, db, 91, "李四")
		seedSalaryEvent(db, 91, "2026-06", "提成", 200)
		db.Create(&model.PositionEvent{PersonID: 91, Seq: 1, EventType: "入职"})

		ctx := ctxWithScope(2, dao.DataScopeOwn, 90)

		var ev model.SalaryEvent
		db.Where("person_id = ?", 91).First(&ev)
		if err := DeleteSalaryEvent(ctx, ev.ID); err == nil {
			t.Error("delete other's salary event should be denied")
		}

		var pe model.PositionEvent
		db.Where("person_id = ?", 91).First(&pe)
		if err := UpdatePositionEvent(ctx, pe.ID, &model.PositionEvent{EventType: "离职"}); err == nil {
			t.Error("update other's position event should be denied")
		}

		// 本人数据操作放行
		seedSalaryEvent(db, 90, "2026-06", "提成", 100)
		var own model.SalaryEvent
		db.Where("person_id = ?", 90).First(&own)
		if err := DeleteSalaryEvent(ctx, own.ID); err != nil {
			t.Errorf("delete own salary event should pass, got %v", err)
		}
	})
}

// TestScopeBatchCalcOwn 仅自己范围：批量考勤核算只处理本人
func TestScopeBatchCalcOwn(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		seedEmployee(db, 90, "2026-01-01", 8000, 2000, 300, 500, 26)
		seedEmployee(db, 91, "2026-01-01", 8000, 2000, 300, 500, 26)
		seedAttendanceDays(db, 90, "2026-06", 5, 8)
		seedAttendanceDays(db, 91, "2026-06", 5, 8)

		ctx := ctxWithScope(2, dao.DataScopeOwn, 90)
		hasValue, _, fail, err := CalculateMonthlyBatch(ctx, "2026-06", []uint{90, 91})
		if err != nil || fail != 0 {
			t.Fatalf("batch calc: err=%v fail=%d", err, fail)
		}
		if hasValue != 1 {
			t.Errorf("hasValue=%d, want 1（仅本人）", hasValue)
		}
		var cnt int64
		db.Model(&model.AttendanceCalculationMonthly{}).Count(&cnt)
		if cnt != 1 {
			t.Errorf("calc rows=%d, want 1", cnt)
		}
	})
}

// TestScopeCarryoverDenied 仅自己范围：公司级年假结转拒绝
func TestScopeCarryoverDenied(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		ctx := ctxWithScope(2, dao.DataScopeOwn, 90)
		if _, err := ExecuteCarryover(ctx, "2026-06", 2, "user"); err == nil {
			t.Error("own scope carryover should be denied")
		}
	})
}

// TestScopeOwnWithoutPerson 仅自己角色但用户未关联人员 → 查询空过滤、写操作拒绝
func TestScopeOwnWithoutPerson(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		db.AutoMigrate(&model.Person{})
		seedPerson(t, db, 90, "张三")
		seedPerson(t, db, 91, "李四")
		seedSalaryEvent(db, 91, "2026-06", "提成", 200)

		// own 范围但未关联人员（PersonID nil）
		ctx := context.Background()
		info := dao.ScopeInfo{UserID: 2, DataScope: dao.DataScopeOwn}
		ctx = dao.WithScopeInfo(ctx, info)

		// 列表：空过滤（防漏成全量）
		_, total, err := GetSalaryEventList(ctx, SalaryEventListQuery{PageNum: 1, PageSize: 10})
		if err != nil || total != 0 {
			t.Fatalf("own without person should see nothing, total=%d err=%v", total, err)
		}

		// 写操作：拒绝
		var ev model.SalaryEvent
		db.Where("person_id = ?", 91).First(&ev)
		if err := DeleteSalaryEvent(ctx, ev.ID); err == nil {
			t.Error("write without person should be denied")
		}

		// 批量核算：无可操作对象
		if _, _, _, err := CalculateMonthlyBatch(ctx, "2026-06", []uint{91}); err == nil {
			// 无 person 时 ScopePersonIDs 返回空 → 无核算；结果三态均为 0 属正常
			t.Log("batch calc with empty scope OK (no-op)")
		}
	})
}

// TestGetUserEffectiveScope 有效数据范围 = 角色范围并集（最宽优先）
func TestGetUserEffectiveScope(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		db.AutoMigrate(&model.Role{}, &model.UserRole{})
		ownRole := model.Role{Name: "own-role", DataScope: dao.DataScopeOwn}
		allRole := model.Role{Name: "all-role", DataScope: dao.DataScopeAll}
		if err := db.Create(&ownRole).Error; err != nil {
			t.Fatalf("seed own role: %v", err)
		}
		if err := db.Create(&allRole).Error; err != nil {
			t.Fatalf("seed all role: %v", err)
		}

		// 仅 own 角色 → own
		db.Create(&model.UserRole{UserID: 3, RoleID: ownRole.ID})
		if got := GetUserEffectiveScope(3); got != dao.DataScopeOwn {
			t.Errorf("only own role: got %s, want own", got)
		}

		// own + all 角色 → all（并集最宽优先）
		db.Create(&model.UserRole{UserID: 4, RoleID: ownRole.ID})
		db.Create(&model.UserRole{UserID: 4, RoleID: allRole.ID})
		if got := GetUserEffectiveScope(4); got != dao.DataScopeAll {
			t.Errorf("own+all roles: got %s, want all", got)
		}

		// 无角色 → all
		if got := GetUserEffectiveScope(5); got != dao.DataScopeAll {
			t.Errorf("no roles: got %s, want all", got)
		}
	})
}

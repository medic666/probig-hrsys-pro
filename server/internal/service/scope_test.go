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

// TestScopeCreateUserValidation 数据范围为「仅自己」时必须关联人员
func TestScopeCreateUserValidation(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		db.AutoMigrate(&model.User{})
		if _, err := CreateUser(context.Background(), CreateUserReq{
			Username: "u1", Password: "x", DataScope: dao.DataScopeOwn,
		}); err == nil {
			t.Error("own scope without person should be rejected")
		}
		pid := uint(90)
		if _, err := CreateUser(context.Background(), CreateUserReq{
			Username: "u2", Password: "x", DataScope: dao.DataScopeOwn, PersonID: &pid,
		}); err != nil {
			t.Errorf("own scope with person should pass, got %v", err)
		}
	})
}

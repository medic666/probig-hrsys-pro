package service

import (
	"testing"

	"probig/server/internal/model"
	"probig/server/internal/utils"

	"gorm.io/gorm"
)

func seedPerson(t *testing.T, db *gorm.DB, id uint, name string) {
	t.Helper()
	if err := db.Create(&model.Person{ID: id, Name: name}).Error; err != nil {
		t.Fatalf("seed person: %v", err)
	}
}

func seedALEvent(t *testing.T, db *gorm.DB, personID uint, evType, date string, hours float64) {
	t.Helper()
	d, _ := utils.ParseDate(date)
	var maxSeq int
	db.Unscoped().Model(&model.AnnualLeaveAccountEvent{}).Where("person_id = ?", personID).
		Select("COALESCE(MAX(seq), 0)").Scan(&maxSeq)
	if err := db.Create(&model.AnnualLeaveAccountEvent{
		PersonID: personID, Seq: maxSeq + 1, EventType: evType,
		SourceType: "manual", Hours: hours, EffectiveDate: utils.DateOnlyFromTime(d),
	}).Error; err != nil {
		t.Fatalf("seed al event: %v", err)
	}
}

func seedAttendanceLeave(t *testing.T, db *gorm.DB, personID uint, date string, hours float64) {
	t.Helper()
	d, _ := utils.ParseDate(date)
	dOnly := utils.DateOnlyFromTime(d)
	daily := model.AttendanceDaily{PersonID: personID, EventDate: dOnly, Status: "confirmed"}
	if err := db.Create(&daily).Error; err != nil {
		t.Fatalf("seed daily: %v", err)
	}
	if err := db.Create(&model.AttendanceEventDetail{DailyID: daily.ID, EventType: "休假", SubType: "年假", Hours: hours}).Error; err != nil {
		t.Fatalf("seed detail: %v", err)
	}
}

func TestAnnualLeaveEventListMergesAttendance(t *testing.T) {
	withSalaryDB(t, func(db *gorm.DB) {
		if err := db.AutoMigrate(&model.Person{}); err != nil {
			t.Fatalf("migrate: %v", err)
		}
		seedPerson(t, db, 80, "张三")
		seedALEvent(t, db, 80, "grant", "2026-01-01", 40)
		seedALEvent(t, db, 80, "carryover_deduct", "2026-12-31", 8)
		seedAttendanceLeave(t, db, 80, "2026-06-10", 8)

		// 全量合并：total=3，按日期倒序
		list, total, err := GetAnnualLeaveEventList(1, 10, 0, "", "", "")
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if total != 3 {
			t.Fatalf("total: got %d, want 3", total)
		}
		wantOrder := []string{"2026-12-31", "2026-06-10", "2026-01-01"}
		for i, w := range wantOrder {
			got := list[i]["effective_date"].(utils.DateOnly).String()
			if got != w {
				t.Errorf("order[%d]: got %s, want %s", i, got, w)
			}
		}
		// attendance 记录 source_type 正确且带姓名
		att := list[1]
		if att["source_type"] != "attendance" {
			t.Errorf("source_type: got %v, want attendance", att["source_type"])
		}
		if att["person_name"] != "张三" {
			t.Errorf("person_name: got %v, want 张三", att["person_name"])
		}
		if att["event_type"] != "休假" {
			t.Errorf("event_type: got %v, want 休假", att["event_type"])
		}

		// 分页：pageSize=2 只返回前两条
		list2, total2, _ := GetAnnualLeaveEventList(1, 2, 0, "", "", "")
		if total2 != 3 || len(list2) != 2 {
			t.Fatalf("pagination: total=%d len=%d, want 3/2", total2, len(list2))
		}

		// 人员过滤
		_, total3, _ := GetAnnualLeaveEventList(1, 10, 80, "", "", "")
		if total3 != 3 {
			t.Errorf("person filter: total=%d, want 3", total3)
		}
		_, total4, _ := GetAnnualLeaveEventList(1, 10, 999, "", "", "")
		if total4 != 0 {
			t.Errorf("person filter none: total=%d, want 0", total4)
		}

		// 类型过滤：grant 只回 account 段
		_, total5, _ := GetAnnualLeaveEventList(1, 10, 0, "", "", "grant")
		if total5 != 1 {
			t.Errorf("type filter grant: total=%d, want 1", total5)
		}
		// 类型过滤：休假 只回 attendance 段
		_, total6, _ := GetAnnualLeaveEventList(1, 10, 0, "", "", "休假")
		if total6 != 1 {
			t.Errorf("type filter 休假: total=%d, want 1", total6)
		}

		// 日期范围过滤
		_, total7, _ := GetAnnualLeaveEventList(1, 10, 0, "2026-06-01", "2026-06-30", "")
		if total7 != 1 {
			t.Errorf("date range filter: total=%d, want 1", total7)
		}

		// attendance 记录只读：不写入 account 事件表
		var acctCount int64
		db.Model(&model.AnnualLeaveAccountEvent{}).Count(&acctCount)
		if acctCount != 2 {
			t.Errorf("account events should remain 2, got %d", acctCount)
		}
	})
}

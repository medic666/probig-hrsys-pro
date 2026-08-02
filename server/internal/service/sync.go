package service

import "gorm.io/gorm"

// SyncChildRecords 通用子表同步（UPSERT）：
// 按主键逐条对比——未变化零操作（零审计）；变化 → 更新（一条"修改"审计，含前后快照）；
// 新出现 → 新增；旧记录不在新列表中 → 软删除。事务由调用方包裹。
//
// 参数：
//
//	tx        事务
//	parentCol 父级外键列名（如 person_id / daily_id）
//	parentID  父级主键
//	incoming  新提交的子记录列表（含 id 的行按主键匹配；id=0 视为新增）
//	keyOf     取记录主键
//	equal     比较内容是否相同（相同则不操作）
//	setParent 写入前设置父级外键
func SyncChildRecords[T any](
	tx *gorm.DB,
	parentCol string,
	parentID uint,
	incoming []T,
	keyOf func(T) uint,
	equal func(a, b T) bool,
	setParent func(*T),
) error {
	var old []T
	if err := tx.Where(parentCol+" = ?", parentID).Find(&old).Error; err != nil {
		return err
	}
	oldMap := make(map[uint]T, len(old))
	for _, o := range old {
		oldMap[keyOf(o)] = o
	}
	seen := make(map[uint]bool, len(incoming))
	for _, in := range incoming {
		key := keyOf(in)
		if key == 0 {
			setParent(&in)
			if err := tx.Create(&in).Error; err != nil {
				return err
			}
			continue
		}
		seen[key] = true
		o, ok := oldMap[key]
		if !ok {
			continue // 传入的 id 不属于当前父级，忽略
		}
		if equal(o, in) {
			continue // 未变化，零操作零审计
		}
		// 变化：以新值覆盖旧记录并保存（触发"修改"审计）
		o = in
		setParent(&o)
		if err := tx.Save(&o).Error; err != nil {
			return err
		}
	}
	for _, o := range old {
		if !seen[keyOf(o)] {
			if err := tx.Delete(&o).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

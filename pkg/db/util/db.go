package util

import "gorm.io/gorm"

func AddWhere(db *gorm.DB, conditions []any) *gorm.DB {
	for _, condition := range conditions {
		if condition != nil {
			cond, ok := condition.([]any)
			if ok {
				if len(cond) == 1 {
					db = db.Where(cond[0])
				} else if len(cond) > 1 {
					db = db.Where(cond[0], cond[1:]...)
				}
			} else {
				db = db.Where(condition)
			}
		}
	}
	return db
}

func AddOrder(db *gorm.DB, order ...string) *gorm.DB {
	if len(order) > 0 {
		for _, v := range order {
			db = db.Order(v)
		}
	}
	return db
}

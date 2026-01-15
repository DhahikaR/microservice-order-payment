package helper

import "gorm.io/gorm"

func CommitOrRollback(tx *gorm.DB) {
	if r := recover(); r != nil {
		tx.Rollback()
		panic(r)
	}
	tx.Commit()
}

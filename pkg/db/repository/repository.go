package repository

import (
	"github.com/misakacoder/inuyasha/pkg/db/util"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository[M any] struct {
	DB *gorm.DB
}

func (repository *Repository[M]) Create(model ...*M) error {
	return repository.DB.Create(model).Error
}

func (repository *Repository[M]) CreateOnConflictDoNothing(models ...*M) error {
	return repository.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(models).Error
}

func (repository *Repository[M]) Update(model *M, fields ...string) error {
	return repository.DB.Select(fields).Updates(model).Error
}

func (repository *Repository[M]) Updates(model *M, conditions []any, fields ...string) error {
	return util.AddWhere(repository.DB, conditions).Select(fields).Updates(model).Error
}

func (repository *Repository[M]) Delete(id uint) error {
	var model M
	return repository.DB.Delete(&model, id).Error
}

func (repository *Repository[M]) Deletes(model *M) error {
	var mod M
	return repository.DB.Where(model).Delete(&mod).Error
}

func (repository *Repository[M]) PrimaryKey(id uint) (M, error) {
	var result M
	err := repository.DB.First(&result, id).Error
	return result, err
}

func (repository *Repository[M]) First(model *M) (M, error) {
	var result M
	err := repository.DB.Where(model).First(&result).Error
	return result, err
}

func (repository *Repository[M]) Take(conditions []any, order ...string) (M, error) {
	var result M
	tx := util.AddWhere(repository.DB, conditions)
	err := util.AddOrder(tx, order...).Take(&result).Error
	return result, err
}

func (repository *Repository[M]) Find(model *M, limit int, order ...string) ([]M, error) {
	var result []M
	err := util.AddOrder(repository.DB, order...).Where(model).Limit(limit).Find(&result).Error
	return result, err
}

func (repository *Repository[M]) Finds(conditions []any, limit int, order ...string) ([]M, error) {
	var result []M
	tx := util.AddWhere(repository.DB, conditions)
	err := util.AddOrder(tx, order...).Limit(limit).Find(&result).Error
	return result, err
}

func (repository *Repository[M]) FindAll(model *M, order ...string) ([]M, error) {
	return repository.Find(model, -1, order...)
}

func (repository *Repository[M]) FindsAll(conditions []any, order ...string) ([]M, error) {
	return repository.Finds(conditions, -1, order...)
}

func (repository *Repository[M]) Page(model *M, page *util.Page) util.PageResult[M] {
	return util.Paginate[M](repository.DB, model, page)
}

func (repository *Repository[M]) Pages(conditions []any, page *util.Page) util.PageResult[M] {
	return util.Paginate[M](repository.DB, conditions, page)
}

func (repository *Repository[M]) Count(model *M) (int64, error) {
	var mod M
	var count int64
	err := repository.DB.Model(&mod).Where(model).Count(&count).Error
	return count, err
}

func (repository *Repository[M]) Counts(conditions []any) (int64, error) {
	var mod M
	var count int64
	tx := util.AddWhere(repository.DB.Model(&mod), conditions)
	err := tx.Count(&count).Error
	return count, err
}

func (repository *Repository[M]) Transaction(fn func(*Repository[M]) error) error {
	tx := repository.DB.Begin()
	defer func() {
		if err := recover(); err != nil {
			tx.Rollback()
			panic(err)
		}
	}()
	repo := &Repository[M]{DB: tx}
	err := fn(repo)
	if err != nil {
		tx.Rollback()
	} else {
		err = tx.Commit().Error
	}
	return err
}

func NewRepository[M any](db *gorm.DB) *Repository[M] {
	return &Repository[M]{DB: db}
}

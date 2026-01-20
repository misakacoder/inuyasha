package util

import (
	"fmt"
	"github.com/misakacoder/kagome/cond"
	"gorm.io/gorm"
	"math"
)

type Page struct {
	OrderBy  string `form:"orderBy"`
	PageNum  int    `form:"pageNum"`
	PageSize int    `form:"pageSize"`
}

type PageResult[T any] struct {
	PageNum int `json:"pageNum"`
	Pages   int `json:"pages"`
	Total   int `json:"total"`
	List    []T `json:"list"`
}

func Paginate[M any](db *gorm.DB, condition any, page *Page) PageResult[M] {
	return PaginateResult[M, M](db, condition, page)
}

func PaginateResult[M, R any](db *gorm.DB, condition any, page *Page) PageResult[R] {
	var model M
	conditions, ok := condition.([]any)
	if !ok {
		conditions = append(conditions, condition)
	}
	countDB := AddWhere(db.Model(model), conditions)
	queryDB := AddWhere(db.Model(model), conditions)
	return paginate[R](countDB, queryDB, page)
}

func PaginateSQL[R any](db *gorm.DB, sql string, args []any, page *Page) PageResult[R] {
	countDB := db.Raw(fmt.Sprintf("select count(1) from (%s) table_count", sql), args...)
	queryDB := db.Raw(sql, args...)
	return paginate[R](countDB, queryDB, page)
}

func paginate[R any](countDB *gorm.DB, queryDB *gorm.DB, page *Page) PageResult[R] {
	rewritePage(page)
	pageResult := PageResult[R]{
		PageNum: page.PageNum,
		List:    []R{},
	}
	var count int64
	countDB.Count(&count)
	pageResult.Total = int(count)
	pages := int(math.Ceil(float64(count) / float64(page.PageSize)))
	pageResult.Pages = pages
	if count == 0 || pageResult.PageNum > pages {
		return pageResult
	}
	var data []R
	offset := (page.PageNum - 1) * page.PageSize
	queryDB.Order(page.OrderBy).Offset(offset).Limit(page.PageSize).Find(&data)
	pageResult.List = data
	return pageResult
}

func rewritePage(page *Page) {
	pageNum := page.PageNum
	pageSize := page.PageSize
	pageNum = cond.Ternary(pageNum <= 0, 1, pageNum)
	pageSize = cond.Ternary(pageSize <= 0, 10, pageSize)
	page.PageNum = pageNum
	page.PageSize = pageSize
}

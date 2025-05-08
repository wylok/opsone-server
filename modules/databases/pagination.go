package databases

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"math"
)

// Param 分页参数
type Pagination struct {
	DB      *gorm.DB
	Page    int
	PerPage int
}

// Paginator 分页返回
type Paginator struct {
	TotalCount int64 `json:"totalCount"`
	PageSize   int   `json:"pageSize"`
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	PrevPage   int   `json:"prev_page"`
	NextPage   int   `json:"next_page"`
}

// Paging 分页
func (p *Pagination) Paging(r interface{}) (*Paginator, interface{}) {
	defer func() {
		if ro := recover(); ro != nil {
			fmt.Println(errors.New(fmt.Sprint(ro)))
		}
	}()
	var (
		done      = make(chan bool, 1)
		paginator Paginator
		count     int64
		offset    int
		db        = p.DB
	)
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PerPage == 0 {
		p.PerPage = 10
	}
	go countRecords(db, r, done, &count)
	if p.Page == 1 {
		offset = 0
	} else {
		offset = (p.Page - 1) * p.PerPage
	}
	<-done
	db.Limit(p.PerPage).Offset(offset).Find(r)
	paginator.TotalCount = count
	paginator.Page = p.Page
	paginator.PerPage = p.PerPage
	paginator.PageSize = int(math.Ceil(float64(count) / float64(p.PerPage)))
	if paginator.PageSize < 1 {
		paginator.PageSize = 1
	}
	if p.Page > 1 {
		paginator.PrevPage = p.Page - 1
	} else {
		paginator.PrevPage = p.Page
	}
	if p.Page == paginator.PageSize {
		paginator.NextPage = p.Page
	} else {
		paginator.NextPage = p.Page + 1
	}
	return &paginator, r
}
func countRecords(db *gorm.DB, anyType interface{}, done chan bool, count *int64) {
	db.Model(anyType).Count(count)
	done <- true
}

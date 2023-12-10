package helper

import (
	"encoding/json"
	"math"

	"github.com/Pheethy/sqlx"
)

const (
	PSQL_TOTAL_ROW_KEY = "total_row"
)

type Paginator struct {
	Page            int `json:"page"`
	PerPage         int `json:"per_page"`
	TotalPages      int `json:"total_page"`
	TotalEntrySizes int `json:"total_rows"`
}

func (p Paginator) String() string {
	ju, _ := json.Marshal(p)
	return string(ju)
}

func NewPaginator() Paginator {
	return Paginator{Page: 1, PerPage: 20}
}

func NewPaginatorWithParams(page int, perPage int) Paginator {
	return Paginator{Page: page, PerPage: perPage}
}

func (p *Paginator) SetPaginatorByAllRows(allRows int) {
	p.setTotalEntrySizes(allRows)
	p.setTotalPages()
}

func (p *Paginator) setTotalEntrySizes(allRows int) {
	p.TotalEntrySizes = allRows
}

func (p *Paginator) setTotalPages() {
	totalRows := p.TotalEntrySizes
	perPage := p.PerPage
	totalPage := math.Ceil(float64(totalRows) / float64(perPage))
	p.TotalPages = int(totalPage)
}

func (p *Paginator) SetTotalFromRows(rows *sqlx.Rows) error {
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	values, err := rows.SliceScan()
	if err != nil {
		return err
	}
	if columns != nil && values != nil {
		if len(columns) > 0 && len(values) > 0 {
			for index, column := range columns {
				if column == PSQL_TOTAL_ROW_KEY {
					total := int(values[index].(int64))
					p.setTotalEntrySizes(total)
					p.setTotalPages()
				}
			}
		}
	}
	return nil
}

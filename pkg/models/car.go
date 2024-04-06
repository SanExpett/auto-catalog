package models

import (
	"github.com/asaskevich/govalidator"
	"github.com/microcosm-cc/bluemonday"
	"strings"
	"time"
)

const (
	minCarYear = 1885
)

func init() {
	govalidator.CustomTypeTagMap.Set("regNumCheck", func(i interface{}, o interface{}) bool {
		if regNum, ok := i.(string); ok {
			regex := `^[A-Za-zА-Яа-я]{1}\d{3}[A-Za-zА-Яа-я]{2}\d{3}$`

			return govalidator.StringMatches(regNum, regex)
		}

		return false
	})

	govalidator.CustomTypeTagMap.Set("yearCheck", func(i interface{}, o interface{}) bool {
		if i == nil {
			return true
		}

		if year, ok := i.(uint64); ok {
			return year > minCarYear
		}

		return false
	})
}

type Car struct {
	ID        uint64    `json:"id"          valid:"required"`
	OwnerID   uint64    `json:"owner_id"    valid:"required"`
	RegNum    string    `json:"reg_num"     valid:"required,regNumCheck"`
	Mark      string    `json:"mark"        valid:"required"`
	Model     string    `json:"model"       valid:"required"`
	Year      uint64    `json:"year"        valid:"optional,yearCheck"`
	CreatedAt time.Time `json:"created_at"  valid:"required"`
}

type PreCar struct {
	OwnerID uint64 `json:"owner_id"    valid:"required"`
	RegNum  string `json:"reg_num"     valid:"required,regNumCheck"`
	Mark    string `json:"mark"        valid:"required"`
	Model   string `json:"model"       valid:"required"`
	Year    uint64 `json:"year"        valid:"optional,yearCheck"`
}

func (c *Car) Sanitize() {
	sanitizer := bluemonday.UGCPolicy()

	c.RegNum = sanitizer.Sanitize(c.RegNum)
	c.Mark = sanitizer.Sanitize(c.Mark)
	c.Model = sanitizer.Sanitize(c.Model)
}

func (c *PreCar) Trim() {
	c.RegNum = strings.TrimSpace(c.RegNum)
	c.Mark = strings.TrimSpace(c.Mark)
	c.Model = strings.TrimSpace(c.Model)
}

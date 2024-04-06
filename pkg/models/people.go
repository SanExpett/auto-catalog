package models

import (
	"github.com/microcosm-cc/bluemonday"
	"strings"
	"time"
)

type People struct {
	ID         uint64    `json:"id"          valid:"required"`
	Name       string    `json:"name"        valid:"required"`
	Surname    string    `json:"surname"     valid:"required"`
	Patronymic string    `json:"patronymic"  valid:"optional"`
	CreatedAt  time.Time `json:"created_at"  valid:"required"`
}

type PrePeople struct {
	Name       string `json:"name"        valid:"required"`
	Surname    string `json:"surname"     valid:"required"`
	Patronymic string `json:"patronymic"  valid:"optional"`
}

func (p *PrePeople) Trim() {
	p.Name = strings.TrimSpace(p.Name)
	p.Surname = strings.TrimSpace(p.Surname)
	p.Patronymic = strings.TrimSpace(p.Patronymic)
}

func (p *People) Sanitize() {
	sanitizer := bluemonday.UGCPolicy()

	p.Name = sanitizer.Sanitize(p.Name)
	p.Surname = sanitizer.Sanitize(p.Surname)
	p.Patronymic = sanitizer.Sanitize(p.Patronymic)
}

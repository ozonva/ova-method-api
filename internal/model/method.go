package model

import (
	"fmt"
	"time"
)

type Method struct {
	Id        uint64    `db:"id"`
	UserId    uint64    `db:"user_id"`
	Value     string    `db:"value"`
	CreatedAt time.Time `db:"created_at"`
}

func (m *Method) String() string {
	return fmt.Sprintf(
		"id[%d], userId[%d], value[%s], created at[%s]",
		m.Id,
		m.UserId,
		m.Value,
		m.CreatedAt.String(),
	)
}

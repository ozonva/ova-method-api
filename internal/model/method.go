package model

import "fmt"

type Method struct {
	Id        uint64
	UserId    uint64
	Value     string
	CreatedAt int64
}

func (m *Method) String() string {
	return fmt.Sprintf(
		"id[%d], userId[%d], value[%s], created at[%d]",
		m.Id,
		m.UserId,
		m.Value,
		m.CreatedAt,
	)
}

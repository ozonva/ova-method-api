package model

import "fmt"

type Method struct {
	UserId    uint64
	Value     string
	CreatedAt int64
}

func (m *Method) String() string {
	return fmt.Sprintf(
		"userId[%d], value[%s], created at[%d]",
		m.UserId,
		m.Value,
		m.CreatedAt,
	)
}

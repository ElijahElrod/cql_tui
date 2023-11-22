package tui

import "encoding/json"

type CT string

const (
	VARCHAR = "string"
)

type CqlType interface {
	Print(tableWidth int) string
}

type CqlString struct {
	CqlType         CT
	Data            string
	PrettyPrintJson bool
}

func (cs *CqlString) Print(table_width int) string {
	if cs.PrettyPrintJson && json.Valid([]byte(cs.Data)) {
		var data interface{}
		err := json.Unmarshal([]byte(cs.Data), &data)
		if err != nil {
			return ""
		}
		out, err := json.MarshalIndent(data, "", "    ")
		if err != nil {
			return cs.Data
		}
		return string(out)
	}
	return cs.Data
}

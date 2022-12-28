package cobrafile

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type PersonnelList []Person

type Person struct {
	Number         int
	WorkerNumber   string
	Name           string
	CardID         string
	Department     string
	Attendance     bool
	AccessControl  bool
	OtherShift     bool
	DeactivateDate time.Time
}

func LoadPersonnel(filename string) (PersonnelList, error) {
	contents, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	reader := csv.NewReader(bytes.NewReader(contents))
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("missing header row")
	}
	headerRow := rows[0]
	rows = rows[1:]

	for c, value := range headerRow {
		value = strings.ToLower(value)
		for strings.HasSuffix(value, ".") {
			value = strings.TrimSuffix(value, ".")
		}
		headerRow[c] = value
	}

	result := make([]Person, 0, len(rows))
	for r, row := range rows {
		p := Person{}

		for c, value := range row {
			switch headerRow[c] {
			case "no":
				v, err := strconv.ParseInt(value, 10, 32)
				if err != nil {
					return nil, fmt.Errorf("row %d: could not parse number: %w", r, err)
				}
				p.Number = int(v)
			case "worker no":
				p.WorkerNumber = value
			case "card id":
				p.CardID = value
			case "department":
				p.Department = value
			case "attendance":
				p.Attendance = (value == "1")
			case "access control":
				p.AccessControl = (value == "1")
			case "other shift":
				p.OtherShift = (value == "1")
			case "deactive":
				v, err := time.Parse("1/2/2006", value)
				if err != nil {
					return nil, fmt.Errorf("row %d: could not parse deactive: %w", r, err)
				}
				p.DeactivateDate = v
			}
		}

		result = append(result, p)
	}
	return result, nil
}

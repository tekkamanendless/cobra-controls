package cobrafile

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type ControllerList []Controller

type Controller struct {
	Name    string
	Address string
	Port    int
	SN      string
	Doors   []string
}

func LoadController(filename string) (ControllerList, error) {
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

	result := make([]Controller, 0, len(rows))
	for r, row := range rows {
		p := Controller{
			Doors: make([]string, 4),
		}

		for c, value := range row {
			switch headerRow[c] {
			case "name":
				p.Name = value
			case "address":
				p.Address = value
			case "port":
				v, err := strconv.ParseInt(value, 10, 32)
				if err != nil {
					return nil, fmt.Errorf("row %d: could not parse port: %w", r, err)
				}
				p.Port = int(v)
			case "sn":
				p.SN = value
			case "door 1":
				p.Doors[0] = value
			case "door 2":
				p.Doors[1] = value
			case "door 3":
				p.Doors[2] = value
			case "door 4":
				p.Doors[3] = value
			}
		}

		result = append(result, p)
	}
	return result, nil
}
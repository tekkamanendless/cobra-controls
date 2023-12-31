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
	Port    uint16
	SN      uint16
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
				v, err := strconv.ParseInt(value, 10, 17) // 16-bit port, plus one for sign.
				if err != nil {
					return nil, fmt.Errorf("row %d: could not parse port: %w", r, err)
				}
				p.Port = uint16(v)
			case "sn":
				v, err := strconv.ParseInt(value, 10, 17) // 16-bit address, plus one for sign.
				if err != nil {
					return nil, fmt.Errorf("row %d: could not parse port: %w", r, err)
				}
				p.SN = uint16(v)
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

// LookupName returns the controller name.
//
// If no controller is found, this returns the empty string.
func (l ControllerList) LookupName(address string) string {
	for _, controller := range l {
		if controller.Address == address {
			return controller.Name
		}
	}
	return ""
}

// FindDoor returns the 1-index door number.
func (l ControllerList) FindDoor(address string, door string) (uint8, bool) {
	for _, controller := range l {
		if controller.Address == address {
			for d := range controller.Doors {
				if strings.EqualFold(controller.Doors[d], door) {
					return uint8(d) + 1, true
				}
			}
		}
	}
	return 0, false
}

// LookupDoor returns the 1-index door number.
func (l ControllerList) LookupDoor(address string, door uint8) string {
	for _, controller := range l {
		if controller.Address == address {
			if door > 0 && int(door) <= len(controller.Doors) {
				return controller.Doors[door-1]
			}
		}
	}
	return ""
}

// LookupNameAndDoor returns the controller name and the 1-index door number.
func (l ControllerList) LookupNameAndDoor(address string, door uint8) (string, string) {
	for _, controller := range l {
		if controller.Address == address {
			if door > 0 && int(door) <= len(controller.Doors) {
				return controller.Name, controller.Doors[door-1]
			}
			return controller.Name, ""
		}
	}
	return "", ""
}

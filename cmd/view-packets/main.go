package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/sirupsen/logrus"
)

func main() {
	filterPacket := func(packet gopacket.Packet) bool {
		if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
			logrus.Debugf("This is a TCP packet.")
			tcp, _ := tcpLayer.(*layers.TCP)
			logrus.Debugf("From src port %d to dst port %d\n", tcp.SrcPort, tcp.DstPort)
			if tcp.SrcPort != 60000 && tcp.DstPort != 60000 {
				return false
			}
		} else {
			return false
		}
		if len(packet.TransportLayer().LayerPayload()) == 0 {
			return false
		}
		return true
	}

	filenames := os.Args[1:]
	for _, filename := range filenames {
		logrus.Infof("File: %s", filename)
		handle, err := pcap.OpenOffline(filename)
		if err != nil {
			logrus.Errorf("Error opening file: [%T] %v", err, err)
			continue
		}
		packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
		for packet := range packetSource.Packets() {
			if !filterPacket(packet) {
				continue
			}
			logrus.Infof("--------------------")

			fromClient := false
			if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
				if tcp, ok := tcpLayer.(*layers.TCP); ok {
					if tcp.DstPort == 60000 {
						fromClient = true
					}
				}
			}

			data := packet.TransportLayer().LayerPayload()
			logrus.Debugf("Data (%d): %X", len(data), data)

			parseData(data, fromClient)
		}
	}
}

func parseData(data []byte, fromClient bool) error {
	if fromClient {
		logrus.Infof("Mode: Client")
	} else {
		logrus.Infof("Mode: Server")
	}

	if len(data) < 8 {
		return fmt.Errorf("invalid length: %d", len(data))
	}
	startByte := data[0]
	endByte := data[len(data)-1]
	logrus.Debugf("Start byte: %X", startByte)
	logrus.Debugf("End byte: %X", endByte)
	if startByte != 0x7E {
		return fmt.Errorf("invalid start byte: %X (expected: 7E)", startByte)
	}
	if endByte != 0x0D {
		return fmt.Errorf("invalid end byte: %X (expected: 0D)", endByte)
	}

	checksumBytes := data[len(data)-3 : len(data)-1]
	checksum := int(uint16(checksumBytes[1])<<8 | uint16(checksumBytes[0]))
	logrus.Debugf("Expected checksum: %d", checksum)

	data = data[1 : len(data)-3]

	actualChecksum := 0
	for i := 0; i < len(data); i++ {
		actualChecksum += int(data[i])
	}
	logrus.Debugf("Actual checksum: %d", actualChecksum)
	if actualChecksum != checksum {
		return fmt.Errorf("checksum does match: %d (expected: %d)", actualChecksum, checksum)
	}

	boardAddress := []byte{data[1], data[0]}
	data = data[2:]
	functionType := []byte{data[1], data[0]}
	data = data[2:]

	logrus.Infof("Packet:")
	logrus.Infof("Board address: %X", boardAddress)
	logrus.Infof("Function type: %X", functionType)
	logrus.Infof("Remaining data: (%d) %X", len(data), data)

	switch fmt.Sprintf("%X", functionType) {
	case "1081":
		logrus.Infof("Function: Read Operation Status Information")
		if fromClient {
			recordIndex := parseUint32(data[0:4]) // 0x0 and 0xFFFFFFFF mean "latest".
			data = data[4:]
			if !isAll(data, 0) {
				logrus.Warnf("Unexpected data; should be all zeros: %X", data)
			}
			logrus.Infof("Record index: %d", recordIndex)
		} else {
			year := uint8(data[0])
			data = data[1:]
			month := uint8(data[0])
			data = data[1:]
			day := uint8(data[0])
			data = data[1:]
			week := uint8(data[0])
			data = data[1:]
			hour := uint8(data[0])
			data = data[1:]
			minute := uint8(data[0])
			data = data[1:]
			second := uint8(data[0])
			data = data[1:]
			cardRecord := parseUint24(data[0:3])
			data = data[3:]
			popedomAmount := parseUint16(data[0:2])
			data = data[2:]
			indexLocation := data[0:8]
			data = data[8:]
			relayStatus := uint8(data[0])
			data = data[1:]
			doorMagnetButtonState := uint8(data[0])
			data = data[1:]
			reserved1 := uint8(data[0])
			data = data[1:]
			faultNumber := uint8(data[0])
			data = data[1:]
			reserved2 := uint8(data[0])
			data = data[1:]
			reserved3 := uint8(data[0])
			data = data[1:]
			// All remaining bytes are reserved.
			if len(data) != 0 {
				logrus.Warnf("Unexpected remaining data: %X", data)
			}

			logrus.Infof("Current time: %04d-%02d-%0d, week %d, %02d:%02d:%02d", year, month, day, week, hour, minute, second)
			logrus.Infof("Card record: %d", cardRecord)
			logrus.Infof("Popedom amount: %d", popedomAmount)
			logrus.Infof("Index location: %X", indexLocation)
			logrus.Infof("Relay status: %d", relayStatus)
			logrus.Infof("Door magnet button state: %d", doorMagnetButtonState)
			logrus.Infof("Reserved1: %d", reserved1)
			logrus.Infof("Fault number: %d", faultNumber)
			logrus.Infof("Reserved2: %d", reserved2)
			logrus.Infof("Reserved3: %d", reserved3)
		}
	case "1082":
		logrus.Infof("Function: Set basic information")
		if fromClient {
			if !isAll(data, 0) {
				logrus.Warnf("Unexpected data; should be all zeros: %X", data)
			}
		} else {
			// TODO: This appears to be wrong.
			year := uint8(data[0])
			data = data[1:]
			month := uint8(data[0])
			data = data[1:]
			day := uint8(data[0])
			data = data[1:]
			version := uint8(data[0])
			data = data[1:]
			model := uint8(data[0])
			data = data[1:]
			// All remaining bytes are reserved.
			if !isAll(data, 0) {
				logrus.Warnf("Unexpected remaining data; should be all zeros: %X", data)
			}

			logrus.Infof("Drive issuance date: %d-%02d-%02d", year, month, day)
			logrus.Infof("Version: %d", version)
			logrus.Infof("Model: %d", model)
		}
	case "108B":
		logrus.Infof("Function: Set the time")
		logrus.Warnf("TODO NOT IMPLEMENTED")
	case "108D":
		logrus.Infof("Function: Read the records information (by index)")
		logrus.Warnf("TODO NOT IMPLEMENTED")
	case "108E":
		logrus.Infof("Function: Remove a specified number of records")
		logrus.Warnf("TODO NOT IMPLEMENTED")
	case "108F":
		logrus.Infof("Function: Set door controls (online/delay)")
		logrus.Warnf("TODO NOT IMPLEMENTED")
	case "1091":
		logrus.Infof("Function: Upload the mission")
		logrus.Warnf("TODO NOT IMPLEMENTED")
	case "1093":
		logrus.Infof("Function: Clear popedom")
		if fromClient {
			if !isAll(data, 0) {
				logrus.Warnf("Unexpected remaining data; should be all zeros: %X", data)
			}
		} else {
			result := uint8(data[0])
			data = data[1:]
			if !isAll(data, 0) {
				logrus.Warnf("Unexpected remaining data; should be all zeros: %X", data)
			}
			logrus.Infof("Result: %d", result)
		}
	case "1095":
		logrus.Infof("Function: Read popedom")
		logrus.Warnf("TODO NOT IMPLEMENTED")
	case "1096":
		logrus.Infof("Function: Read control period of time")
		logrus.Warnf("TODO NOT IMPLEMENTED")
	case "1097":
		logrus.Infof("Function: Modification control period of time")
		// TODO: This supposedly returns something "from 0" (maybe the period of time index?) on failure; otherwise it returns the same information.
		if fromClient || !fromClient {
			periodOfTimeIndex := parseUint16(data[0:2])
			data = data[2:]
			weekControl := uint8(data[0])
			data = data[1:]
			linkPeriodOfTime := uint8(data[0])
			data = data[1:]
			standby1 := uint8(data[0])
			data = data[1:]
			standby2 := uint8(data[0])
			data = data[1:]
			startTime1, err := parseTime(data[0:2])
			if err != nil {
				logrus.Warnf("Could not parse start time: [%T] %v", err, err)
			}
			data = data[2:]
			endTime1, err := parseTime(data[0:2])
			if err != nil {
				logrus.Warnf("Could not parse end time: [%T] %v", err, err)
			}
			data = data[2:]
			startTime2, err := parseTime(data[0:2])
			if err != nil {
				logrus.Warnf("Could not parse start time: [%T] %v", err, err)
			}
			data = data[2:]
			endTime2, err := parseTime(data[0:2])
			if err != nil {
				logrus.Warnf("Could not parse end time: [%T] %v", err, err)
			}
			data = data[2:]
			startTime3, err := parseTime(data[0:2])
			if err != nil {
				logrus.Warnf("Could not parse start time: [%T] %v", err, err)
			}
			data = data[2:]
			endTime3, err := parseTime(data[0:2])
			if err != nil {
				logrus.Warnf("Could not parse end time: [%T] %v", err, err)
			}
			data = data[2:]
			startTime4, err := parseTime(data[0:2])
			if err != nil {
				logrus.Warnf("Could not parse start time: [%T] %v", err, err)
			}
			data = data[2:]
			endTime4, err := parseTime(data[0:2])
			if err != nil {
				logrus.Warnf("Could not parse end time: [%T] %v", err, err)
			}
			data = data[2:]
			standby := data[0:4]
			data = data[4:]
			if len(data) != 0 {
				logrus.Warnf("Unexpected additional data: (%d)", len(data))
			}

			logrus.Infof("Period of time index: %d", periodOfTimeIndex)
			logrus.Infof("Week control: %d", weekControl)
			logrus.Infof("Link period of time: %d", linkPeriodOfTime)
			logrus.Infof("Standby 1: %d", standby1)
			logrus.Infof("Standby 2: %d", standby2)
			logrus.Infof("Start time 1: %v", startTime1)
			logrus.Infof("End time 1: %v", endTime1)
			logrus.Infof("Start time 2: %v", startTime2)
			logrus.Infof("End time 2: %v", endTime2)
			logrus.Infof("Start time 3: %v", startTime3)
			logrus.Infof("End time 3: %v", endTime3)
			logrus.Infof("Start time 4: %v", startTime4)
			logrus.Infof("End time 4: %v", endTime4)
			logrus.Infof("Standby: %X", standby)
		}
	case "1098":
		// TODO: This appears to be some kind of simple thing, maybe a ping/pong action.
		logrus.Infof("Function: Unknown1098")
		if fromClient {
			logrus.Infof("Unknown: %X", data)
		} else {
			result := uint8(data[0])
			data = data[1:]
			if !isAll(data, 0) {
				logrus.Warnf("Unexpected remaining data; should be all zeros: %X", data)
			}
			logrus.Infof("Result: %d", result)
		}
	case "109B":
		logrus.Infof("Function: Tail plus permissions")
		if fromClient {
			popedomIndex := parseUint16(data[0:2])
			data = data[2:]
			id := parseUint16(data[0:2])
			data = data[2:]
			userNumber := uint8(data[0])
			data = data[1:]
			cardID := fmt.Sprintf("%d%05d", userNumber, id)
			doorNumber := uint8(data[0])
			data = data[1:]
			startDate, err := parseDate(data[0:2])
			if err != nil {
				logrus.Warnf("Could not parse start date: [%T] %v", err, err)
			}
			data = data[2:]
			endDate, err := parseDate(data[0:2])
			if err != nil {
				logrus.Warnf("Could not parse end date: [%T] %v", err, err)
			}
			data = data[2:]
			timeValue := data[0]
			data = data[1:]
			password := data[0:3]
			data = data[3:]
			standby := data[0:4]
			data = data[4:]
			if !isAll(data, 0) {
				logrus.Warnf("Unexpected remaining data; should be all zeros: %X", data)
			}

			logrus.Infof("Popedom index: %d", popedomIndex)
			logrus.Infof("ID: %d", id)
			logrus.Infof("User number: %d", userNumber)
			logrus.Infof("Card ID: %s", cardID)
			logrus.Infof("Door number: %d", doorNumber)
			logrus.Infof("Start date: %v", startDate)
			logrus.Infof("End date: %v", endDate)
			logrus.Infof("Time: %X", timeValue)
			logrus.Infof("Password: %X", password)
			logrus.Infof("Standby: %X", standby)
		} else {
			result := uint8(data[0])
			data = data[1:]
			if !isAll(data, 0) {
				logrus.Warnf("Unexpected remaining data; should be all zeros: %X", data)
			}
			logrus.Infof("Result: %d", result)
		}
	case "109D":
		logrus.Infof("Function: Long-distance open door")
		logrus.Warnf("TODO NOT IMPLEMENTED")
	case "10F1":
		logrus.Infof("Function: Read")
		logrus.Warnf("TODO NOT IMPLEMENTED")
	case "10F4":
		logrus.Infof("Function: Amend, Expand, settings")
		if fromClient {
			address := uint8(data[0])
			data = data[1:]
			unknown1 := uint8(data[0])
			data = data[1:]
			value := uint8(data[0])
			data = data[1:]
			if !isAll(data, 0) {
				logrus.Warnf("Unexpected remaining data; should be all zeros: %X", data)
			}

			logrus.Infof("Address: %02X", address)
			logrus.Infof("Unknown1: %d (should probably be 0)", unknown1)
			logrus.Infof("Value: 0b%08s", strconv.FormatInt(int64(value), 2))
		} else {
			result := uint8(data[0])
			data = data[1:]
			if !isAll(data, 0) {
				logrus.Warnf("Unexpected remaining data; should be all zeros: %X", data)
			}
			logrus.Infof("Result: %d", result)
		}
	case "10F5":
		logrus.Infof("Function: Realize timing task")
		logrus.Warnf("TODO NOT IMPLEMENTED")
	case "10F9":
		// TODO: This appears to be the thing that pushes the config up.
		logrus.Infof("Function: Unknown10F9")
		if fromClient {
			unknown1 := data[0:4]
			data = data[4:]
			logrus.Infof("Unknown1: %X", unknown1)
			for len(data) >= 16 {
				popedom := data[0:16]
				data = data[16:]
				logrus.Infof("Popedom: %X", popedom)
				if fmt.Sprintf("%X", popedom) == "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF" {
					logrus.Infof("Skipping bogus popedom.")
					continue
				}
				id := parseUint16(popedom[0:2])
				popedom = popedom[2:]
				area := uint8(popedom[0])
				popedom = popedom[1:]
				cardID := fmt.Sprintf("%d%05d", area, id)
				door := uint8(popedom[0])
				popedom = popedom[1:]
				openDateBytes := popedom[0:2]
				popedom = popedom[2:]
				logrus.Infof("Open date bytes: %X", openDateBytes)
				openDate, err := parseDate(openDateBytes)
				if err != nil {
					logrus.Warnf("Could not parse open date: [%T] %v", err, err)
				}
				closeDateBytes := popedom[0:2]
				popedom = popedom[2:]
				logrus.Infof("Close date bytes: %X", closeDateBytes)
				closeDate, err := parseDate(closeDateBytes)
				if err != nil {
					logrus.Warnf("Could not parse close date: [%T] %v", err, err)
				}
				controlIndex := popedom[0]
				popedom = popedom[1:]
				password := popedom[0:3]
				popedom = popedom[3:]
				standby1 := popedom[0]
				popedom = popedom[1:]
				standby2 := popedom[0]
				popedom = popedom[1:]
				standby3 := popedom[0]
				popedom = popedom[1:]
				standby4 := popedom[0]
				popedom = popedom[1:]
				if len(popedom) != 0 {
					logrus.Warnf("Unexpected extra popedom data: (%d)", len(popedom))
				}
				logrus.Infof("ID: %d", id)
				logrus.Infof("Area: %d", area)
				logrus.Infof("Card ID: %s", cardID)
				logrus.Infof("Door: %d", door)
				logrus.Infof("Open Date: %v", openDate)
				logrus.Infof("Close Date: %v", closeDate)
				logrus.Infof("Control index: %X", controlIndex) // 0 to not use control time; 1 to specify a time.
				logrus.Infof("Password: %X", password)
				logrus.Infof("Standby 1: %X", standby1) // 1 for the "first card users"; 0 for not those users.
				logrus.Infof("Standby 2: %X", standby2) // 0 for the general user group; >0 for special group permissions.
				logrus.Infof("Standby 3: %X", standby3)
				logrus.Infof("Standby 4: %X", standby4)
			}
			if len(data) != 0 {
				logrus.Warnf("Unexpected trailing data length: (%d)", len(data))
			}
		} else {
			result := uint8(data[0])
			data = data[1:]
			if !isAll(data, 0) {
				logrus.Warnf("Unexpected remaining data; should be all zeros: %X", data)
			}
			logrus.Infof("Result: %d", result)
		}
	case "10FF":
		logrus.Infof("Function: Formatting")
		// This will factory-reset the unit.
		logrus.Warnf("TODO NOT IMPLEMENTED")
	case "1101":
		logrus.Infof("Function: Search .net equipment")
		if fromClient {
			// TODO: We seem to send "1" as the first byte.
			if !isAll(data, 0) {
				logrus.Warnf("Unexpected data; should be all zeros: %X", data)
			}
		} else {
			macAddress := data[0:6]
			data = data[6:]
			ipAddress := data[0:4]
			data = data[4:]
			netmask := data[0:4]
			data = data[4:]
			gateway := data[0:4]
			data = data[4:]
			port := parseUint16(data[0:2])
			data = data[2:]
			// All remaining bytes are reserved.
			if !isAll(data, 0) {
				logrus.Warnf("Unexpected remaining data; should be all zeros: %X", data)
			}

			logrus.Infof("MAC address: %02X:%02X:%02X:%02X:%02X:%02X", macAddress[0], macAddress[1], macAddress[2], macAddress[3], macAddress[4], macAddress[5])
			logrus.Infof("IP address: %d.%d.%d.%d", ipAddress[0], ipAddress[1], ipAddress[2], ipAddress[3])
			logrus.Infof("Netmask: %d.%d.%d.%d", netmask[0], netmask[1], netmask[2], netmask[3])
			logrus.Infof("Gateway: %d.%d.%d.%d", gateway[0], gateway[1], gateway[2], gateway[3])
			logrus.Infof("Port: %d", port)
		}
	case "1107":
		logrus.Infof("Function: Add or modify permissions")
		logrus.Warnf("TODO NOT IMPLEMENTED")
	case "1108":
		logrus.Infof("Function: Delete an authority")
		logrus.Warnf("TODO NOT IMPLEMENTED")
	case "11F2":
		logrus.Infof("Function: Setting TCPIP")
		logrus.Warnf("TODO NOT IMPLEMENTED")
	default:
		logrus.Warnf("Unhandled function: %X", functionType)
	}
	return nil
}

func parseUint16(data []byte) uint16 {
	value := (uint16(data[1]) << 8) | uint16(data[0])
	return value
}

func parseUint24(data []byte) uint32 {
	value := (uint32(data[1]) << 16) | (uint32(data[1]) << 8) | uint32(data[0])
	return value
}

func parseUint32(data []byte) uint32 {
	value := (uint32(data[1]) << 24) | (uint32(data[1]) << 16) | (uint32(data[1]) << 8) | uint32(data[0])
	return value
}

func parseDate(data []byte) (time.Time, error) {
	if len(data) != 2 {
		return time.Time{}, fmt.Errorf("invalid length: %d (expected: 2)", len(data))
	}
	value := parseUint16(data)
	year := (value & 0b1111111000000000) >> 9
	month := (value & 0b0000000111100000) >> 5
	day := (value & 0b0000000000011111) >> 0

	logrus.Debugf("Date: %04d-%02d-%02d", year, month, day)

	output := time.Date(2000+int(year), time.Month(month), int(day), 0, 0, 0, 0, time.UTC)
	return output, nil
}

func parseTime(data []byte) (time.Time, error) {
	if len(data) != 2 {
		return time.Time{}, fmt.Errorf("invalid length: %d (expected: 2)", len(data))
	}
	value := parseUint16(data)
	hours := (value & 0b1111100000000000) >> 11
	minutes := (value & 0b0000011111100000) >> 5
	seconds := (value & 0b0000000000011111) >> 0

	logrus.Debugf("Time: %02d:%02d:%02d", hours, minutes, seconds)

	output := time.Date(0, time.January, 1, int(hours), int(minutes), int(seconds)*2, 0, time.UTC)
	return output, nil
}

func isAll(data []byte, expectedValue byte) bool {
	for _, b := range data {
		if b != expectedValue {
			return false
		}
	}
	return true
}

package main

import (
	"fmt"
	"os"
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
	case "1082":
		logrus.Infof("Function: Set basic information")
	case "108B":
		logrus.Infof("Function: Set the time")
	case "108D":
		logrus.Infof("Function: Read the records information (by index)")
	case "108E":
		logrus.Infof("Function: Remove a specified number of records")
	case "108F":
		logrus.Infof("Function: Set door controls (online/delay)")
	case "1091":
		logrus.Infof("Function: Upload the mission")
	case "1093":
		logrus.Infof("Function: Clear popedom")
	case "1095":
		logrus.Infof("Function: Read popedom")
	case "1096":
		logrus.Infof("Function: Read control period of time")
	case "1097":
		logrus.Infof("Function: Modification control period of time")
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
			logrus.Infof("Result: %d", result)
		}
	case "109D":
		logrus.Infof("Function: Long-distance open door")
	case "10F1":
		logrus.Infof("Function: Read")
	case "10F4":
		logrus.Infof("Function: Amend, Expand, settings")
		// 36: turn off timing mission
	case "10F5":
		logrus.Infof("Function: Realize timing task")
	case "10F9":
		// TODO TODO TOOD
		logrus.Infof("Function: ???")
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
			logrus.Infof("Result: %d", result)
		}
	case "10FF":
		logrus.Infof("Function: Formatting")
	case "1101":
		logrus.Infof("Function: Search .net equipment")
	case "1107":
		logrus.Infof("Function: Add or modify permissions")
	case "1108":
		logrus.Infof("Function: Delete an authority")
	case "11F2":
		logrus.Infof("Function: Setting TCPIP")
	default:
		logrus.Warnf("Unhandled function: %X", functionType)
	}
	return nil
}

func parseUint16(data []byte) uint16 {
	value := uint16((uint16(data[1]) << 8) | uint16(data[0]))
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

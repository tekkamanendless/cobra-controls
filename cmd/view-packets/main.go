package main

import (
	"flag"
	"fmt"
	"strconv"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/sirupsen/logrus"
	"gitlab.com/doug-manley/cobra-controls/wire"
)

func main() {
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	flag.Parse()
	logrus.Infof("verbose: %t", *verbose)
	if *verbose {
		logrus.Infof("Enabling verbose logging.")
		logrus.SetLevel(logrus.DebugLevel)
	}
	filenames := flag.Args()

	filterPacket := func(packet gopacket.Packet) bool {
		if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
			logrus.Debugf("This is a TCP packet.")
			tcp, _ := tcpLayer.(*layers.TCP)
			logrus.Debugf("From src port %d to dst port %d.", tcp.SrcPort, tcp.DstPort)
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

			err = parseData(data, fromClient)
			if err != nil {
				logrus.Warnf("Could not parse data: [%T] %v", err, err)
			}
		}
	}
}

func parseData(fullContents []byte, fromClient bool) error {
	if fromClient {
		logrus.Infof("Mode: Client")
	} else {
		logrus.Infof("Mode: Server")
	}

	var envelope wire.Envelope
	err := wire.Decode(fullContents, &envelope)
	if err != nil {
		return fmt.Errorf("could not decode envelope: %w", err)
	}

	logrus.Infof("Packet:")
	logrus.Infof("Board address: 0x%X", envelope.BoardAddress)
	logrus.Infof("Function type: 0x%X", envelope.Function)
	logrus.Infof("Remaining data: (%d) %X", len(envelope.Contents), envelope.Contents)

	data := wire.NewReader(envelope.Contents)

	switch envelope.Function {
	case wire.FunctionGetOperationStatus:
		logrus.Infof("Function: GetOperationStatus")
		if fromClient {
			var request wire.GetOperationStatusRequest
			err = wire.Decode(data.Bytes(), &request)
			if err != nil {
				return err
			}
			logrus.Infof("Request: %+v", request)
		} else {
			var response wire.GetOperationStatusResponse
			err = wire.Decode(data.Bytes(), &response)
			if err != nil {
				return err
			}
			logrus.Infof("Response: %+v", response)
			if response.Record != nil {
				logrus.Infof("Record: %+v", *response.Record)
			}
		}
	case wire.FunctionGetBasicInfo:
		logrus.Infof("Function: GetBasicInfo")
		if fromClient {
			var request wire.GetBasicInfoRequest
			err = wire.Decode(data.Bytes(), &request)
			if err != nil {
				return err
			}
			logrus.Infof("Request: %+v", request)
		} else {
			var response wire.GetBasicInfoResponse
			err = wire.Decode(data.Bytes(), &response)
			if err != nil {
				return err
			}
			logrus.Infof("Response: %+v", response)
		}
	case 0x108B:
		logrus.Infof("Function: Set the time")
		logrus.Warnf("TODO NOT IMPLEMENTED")
	case 0x108D:
		logrus.Infof("Function: Read the records information (by index)")
		if fromClient {
			recordIndex, err := data.ReadUint32()
			if err != nil {
				return fmt.Errorf("could not read record index: %w", err)
			}
			// All remaining bytes are reserved.
			if !wire.IsAll(data.Bytes(), 0) {
				logrus.Warnf("Unexpected remaining data; should be all zeros: %X", data.Bytes())
			}
			logrus.Infof("Record index: %d", recordIndex)
		} else {
			cardNumber, err := data.ReadUint16()
			if err != nil {
				return fmt.Errorf("could not read card number: %w", err)
			}
			userNumber, err := data.ReadUint8()
			if err != nil {
				return fmt.Errorf("could not read user number: %w", err)
			}
			brushCardState, err := data.ReadUint8()
			if err != nil {
				return fmt.Errorf("could not read brush card state: %w", err)
			}
			brushCardDate, err := data.ReadDate()
			if err != nil {
				return fmt.Errorf("could not read brush card date: %w", err)
			}
			brushCardTime, err := data.ReadTime()
			if err != nil {
				return fmt.Errorf("could not read brush card time: %w", err)
			}
			// All remaining bytes are reserved.
			if !wire.IsAll(data.Bytes(), 0) {
				logrus.Warnf("Unexpected remaining data; should be all zeros: %X", data.Bytes())
			}

			cardID := fmt.Sprintf("%d%05d", userNumber, cardNumber)
			logrus.Infof("   Card number: %d", cardNumber)
			logrus.Infof("   User number: %d", userNumber)
			logrus.Infof("   Card ID: %s", cardID)
			logrus.Infof("   Brush card state: %d", brushCardState)
			logrus.Infof("   Brush date: %v", brushCardDate)
			logrus.Infof("   Brush time: %v", brushCardTime)

		}
	case 0x108E:
		logrus.Infof("Function: Remove a specified number of records")
		if fromClient {
			recordIndex, err := data.ReadUint32()
			if err != nil {
				return fmt.Errorf("could not read record index: %w", err)
			}
			// All remaining bytes are reserved.
			if !wire.IsAll(data.Bytes(), 0) {
				logrus.Warnf("Unexpected remaining data; should be all zeros: %X", data.Bytes())
			}
			logrus.Infof("Record index: %d", recordIndex)
		} else {
			result, err := data.ReadUint8()
			if err != nil {
				return fmt.Errorf("could not read result: %w", err)
			}
			// All remaining bytes are reserved.
			if !wire.IsAll(data.Bytes(), 0) {
				logrus.Warnf("Unexpected remaining data; should be all zeros: %X", data.Bytes())
			}

			logrus.Infof("Result: %d (0 is good)", result)
		}
	case 0x108F:
		logrus.Infof("Function: Set door controls (online/delay)")
		logrus.Warnf("TODO NOT IMPLEMENTED")
	case 0x1091:
		logrus.Infof("Function: Upload the mission")
		logrus.Warnf("TODO NOT IMPLEMENTED")
	case 0x1093:
		logrus.Infof("Function: Clear popedom")
		if fromClient {
			if !wire.IsAll(data.Bytes(), 0) {
				logrus.Warnf("Unexpected remaining data; should be all zeros: %X", data.Bytes())
			}
		} else {
			result, err := data.ReadUint8()
			if err != nil {
				return fmt.Errorf("could not read result: %w", err)
			}
			if !wire.IsAll(data.Bytes(), 0) {
				logrus.Warnf("Unexpected remaining data; should be all zeros: %X", data.Bytes())
			}
			logrus.Infof("Result: %d", result)
		}
	case 0x1095:
		logrus.Infof("Function: Read popedom")
		logrus.Warnf("TODO NOT IMPLEMENTED")
	case 0x1096:
		logrus.Infof("Function: Read control period of time")
		logrus.Warnf("TODO NOT IMPLEMENTED")
	case 0x1097:
		logrus.Infof("Function: Modification control period of time")
		// TODO: This supposedly returns something "from 0" (maybe the period of time index?) on failure; otherwise it returns the same information.
		if fromClient || !fromClient {
			periodOfTimeIndex, err := data.ReadUint16()
			if err != nil {
				return fmt.Errorf("could not read period of time index: %w", err)
			}
			weekControl, err := data.ReadUint8()
			if err != nil {
				return fmt.Errorf("could not read week control: %w", err)
			}
			linkPeriodOfTime, err := data.ReadUint8()
			if err != nil {
				return fmt.Errorf("could not read link period of time: %w", err)
			}
			standby1, err := data.ReadUint8()
			if err != nil {
				return fmt.Errorf("could not read standby 1: %w", err)
			}
			standby2, err := data.ReadUint8()
			if err != nil {
				return fmt.Errorf("could not read standby 2: %w", err)
			}
			startTime1, err := data.ReadTime()
			if err != nil {
				return fmt.Errorf("could not read start time 1: %w", err)
			}
			endTime1, err := data.ReadTime()
			if err != nil {
				return fmt.Errorf("could not read end time 1: %w", err)
			}
			startTime2, err := data.ReadTime()
			if err != nil {
				return fmt.Errorf("could not read start time 2: %w", err)
			}
			endTime2, err := data.ReadTime()
			if err != nil {
				return fmt.Errorf("could not read end time 2: %w", err)
			}
			startTime3, err := data.ReadTime()
			if err != nil {
				return fmt.Errorf("could not read start time 3: %w", err)
			}
			endTime3, err := data.ReadTime()
			if err != nil {
				return fmt.Errorf("could not read end time 3: %w", err)
			}
			startTime4, err := data.ReadTime()
			if err != nil {
				return fmt.Errorf("could not read start time 4: %w", err)
			}
			endTime4, err := data.ReadTime()
			if err != nil {
				return fmt.Errorf("could not read end time 4: %w", err)
			}
			standby, err := data.ReadBytes(4)
			if err != nil {
				return fmt.Errorf("could not read standby: %w", err)
			}
			if data.Length() != 0 {
				logrus.Warnf("Unexpected additional data: (%d)", data.Length())
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
	case 0x1098:
		// TODO: This appears to be some kind of simple thing, maybe a ping/pong action.
		logrus.Infof("Function: Unknown1098")
		if fromClient {
			logrus.Infof("Unknown: %X", data)
		} else {
			result, err := data.ReadUint8()
			if err != nil {
				return fmt.Errorf("could not read result: %w", err)
			}
			if !wire.IsAll(data.Bytes(), 0) {
				logrus.Warnf("Unexpected remaining data; should be all zeros: %X", data.Bytes())
			}
			logrus.Infof("Result: %d", result)
		}
	case 0x109B:
		logrus.Infof("Function: Tail plus permissions")
		if fromClient {
			popedomIndex, err := data.ReadUint16()
			if err != nil {
				return fmt.Errorf("could not read popedom index: %w", err)
			}
			id, err := data.ReadUint16()
			if err != nil {
				return fmt.Errorf("could not read id: %w", err)
			}
			userNumber, err := data.ReadUint8()
			if err != nil {
				return fmt.Errorf("could not read user number: %w", err)
			}
			cardID := fmt.Sprintf("%d%05d", userNumber, id)
			doorNumber, err := data.ReadUint8()
			if err != nil {
				return fmt.Errorf("could not read door number: %w", err)
			}
			startDate, err := data.ReadDate()
			if err != nil {
				return fmt.Errorf("could not read start date: %w", err)
			}
			endDate, err := data.ReadDate()
			if err != nil {
				return fmt.Errorf("could not read end date: %w", err)
			}
			timeValue, err := data.ReadUint8()
			if err != nil {
				return fmt.Errorf("could not read time value: %w", err)
			}
			password, err := data.ReadBytes(3)
			if err != nil {
				return fmt.Errorf("could not read password: %w", err)
			}
			standby, err := data.ReadBytes(4)
			if err != nil {
				return fmt.Errorf("could not read standby: %w", err)
			}
			if !wire.IsAll(data.Bytes(), 0) {
				logrus.Warnf("Unexpected remaining data; should be all zeros: %X", data.Bytes())
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
			result, err := data.ReadUint8()
			if err != nil {
				return fmt.Errorf("could not read result: %w", err)
			}
			if !wire.IsAll(data.Bytes(), 0) {
				logrus.Warnf("Unexpected remaining data; should be all zeros: %X", data.Bytes())
			}
			logrus.Infof("Result: %d", result)
		}
	case 0x109D:
		logrus.Infof("Function: Long-distance open door")
		if fromClient {
			door, err := data.ReadUint8()
			if err != nil {
				return fmt.Errorf("could not read door: %w", err)
			}
			unknown1, err := data.ReadUint8() // I've seen this as "1".
			if err != nil {
				return fmt.Errorf("could not read unknown1: %w", err)
			}
			if !wire.IsAll(data.Bytes(), 0) {
				logrus.Warnf("Unexpected remaining data; should be all zeros: %X", data.Bytes())
			}

			logrus.Infof("Door: %d", door)
			logrus.Infof("Unknown1: %d", unknown1)
		} else {
			if !wire.IsAll(data.Bytes(), 0) {
				logrus.Warnf("Unexpected remaining data; should be all zeros: %X", data.Bytes())
			}
		}
	case 0x10F1:
		logrus.Infof("Function: Read")
		if fromClient {
			address, err := data.ReadUint8()
			if err != nil {
				return fmt.Errorf("could not read address: %w", err)
			}
			reserved, err := data.ReadUint8() // I've seen this as "1".
			if err != nil {
				return fmt.Errorf("could not read reserved: %w", err)
			}
			if !wire.IsAll(data.Bytes(), 0) {
				logrus.Warnf("Unexpected remaining data; should be all zeros: %X", data.Bytes())
			}

			logrus.Infof("Address: %d", address)
			logrus.Infof("Reserved: %d", reserved)
		} else {
			value, err := data.ReadUint8()
			if err != nil {
				return fmt.Errorf("could not read value: %w", err)
			}
			if !wire.IsAll(data.Bytes(), 0) {
				logrus.Warnf("Unexpected remaining data; should be all zeros: %X", data.Bytes())
			}

			logrus.Infof("Value: %d", value)
		}
	case 0x10F4:
		logrus.Infof("Function: Amend, Expand, settings")
		if fromClient {
			address, err := data.ReadUint8()
			if err != nil {
				return fmt.Errorf("could not read address: %w", err)
			}
			unknown1, err := data.ReadUint8()
			if err != nil {
				return fmt.Errorf("could not read unknown1: %w", err)
			}
			value, err := data.ReadUint8()
			if err != nil {
				return fmt.Errorf("could not read value: %w", err)
			}
			if !wire.IsAll(data.Bytes(), 0) {
				logrus.Warnf("Unexpected remaining data; should be all zeros: %X", data.Bytes())
			}

			logrus.Infof("Address: %02X", address)
			logrus.Infof("Unknown1: %d (should probably be 0)", unknown1)
			logrus.Infof("Value: 0b%08s", strconv.FormatInt(int64(value), 2))
		} else {
			result, err := data.ReadUint8()
			if err != nil {
				return fmt.Errorf("could not read result: %w", err)
			}
			if !wire.IsAll(data.Bytes(), 0) {
				logrus.Warnf("Unexpected remaining data; should be all zeros: %X", data.Bytes())
			}
			logrus.Infof("Result: %d", result)
		}
	case 0x10F5:
		logrus.Infof("Function: Realize timing task")
		logrus.Warnf("TODO NOT IMPLEMENTED")
	case 0x10F9:
		// TODO: This appears to be the thing that pushes the config up.
		logrus.Infof("Function: Unknown10F9")
		if fromClient {
			unknown1, err := data.ReadUint8()
			if err != nil {
				return fmt.Errorf("could not read unknown1: %w", err)
			}
			unknown2, err := data.ReadUint8()
			if err != nil {
				return fmt.Errorf("could not read unknown2: %w", err)
			}
			unknown3, err := data.ReadUint8()
			if err != nil {
				return fmt.Errorf("could not read unknown3: %w", err)
			}
			unknown4, err := data.ReadUint8()
			if err != nil {
				return fmt.Errorf("could not read unknown4: %w", err)
			}
			logrus.Infof("Unknown1: %d", unknown1)
			logrus.Infof("Unknown2: %d", unknown2)
			logrus.Infof("Unknown3: %d (index, maybe?)", unknown3)
			logrus.Infof("Unknown4: %d", unknown4)
			for p := 0; data.Length() >= 16; p++ {
				popedom, err := data.Read(16)
				if err != nil {
					return fmt.Errorf("could not read popedom: %w", err)
				}
				logrus.Infof("Popedom[%2d]: %X", p, popedom.Bytes())
				if fmt.Sprintf("%X", popedom.Bytes()) == "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF" {
					logrus.Infof("Skipping bogus popedom.")
					continue
				}
				id, err := popedom.ReadUint16()
				if err != nil {
					return fmt.Errorf("could not read id: %w", err)
				}
				area, err := popedom.ReadUint8()
				if err != nil {
					return fmt.Errorf("could not read area: %w", err)
				}
				cardID := fmt.Sprintf("%d%05d", area, id)
				door, err := popedom.ReadUint8()
				if err != nil {
					return fmt.Errorf("could not read door: %w", err)
				}
				openDate, err := popedom.ReadDate()
				if err != nil {
					return fmt.Errorf("could not read open date: %w", err)
				}
				closeDate, err := popedom.ReadDate()
				if err != nil {
					return fmt.Errorf("could not read close date: %w", err)
				}
				controlIndex, err := popedom.ReadUint8()
				if err != nil {
					return fmt.Errorf("could not read control index: %w", err)
				}
				password, err := popedom.ReadBytes(3)
				if err != nil {
					return fmt.Errorf("could not read password: %w", err)
				}
				standby1, err := popedom.ReadUint8()
				if err != nil {
					return fmt.Errorf("could not read standby 1: %w", err)
				}
				standby2, err := popedom.ReadUint8()
				if err != nil {
					return fmt.Errorf("could not read standby 2: %w", err)
				}
				standby3, err := popedom.ReadUint8()
				if err != nil {
					return fmt.Errorf("could not read standby 3: %w", err)
				}
				standby4, err := popedom.ReadUint8()
				if err != nil {
					return fmt.Errorf("could not read standby 4: %w", err)
				}
				if popedom.Length() != 0 {
					logrus.Warnf("Unexpected extra popedom data: (%d)", popedom.Length())
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
			if data.Length() != 0 {
				logrus.Warnf("Unexpected trailing data length: (%d)", data.Length())
			}
		} else {
			result, err := data.ReadUint8()
			if err != nil {
				return fmt.Errorf("could not read result: %w", err)
			}
			if !wire.IsAll(data.Bytes(), 0) {
				logrus.Warnf("Unexpected remaining data; should be all zeros: %X", data.Bytes())
			}
			logrus.Infof("Result: %d", result)
		}
	case 0x10FF:
		logrus.Infof("Function: Formatting")
		// This will factory-reset the unit.
		logrus.Warnf("TODO NOT IMPLEMENTED")
	case wire.FunctionGetNetworkInfo:
		logrus.Infof("Function: GetNetworkInfo")
		if fromClient {
			var request wire.GetNetworkInfoRequest
			err = wire.Decode(data.Bytes(), &request)
			if err != nil {
				return err
			}
			logrus.Infof("Request: %+v", request)
		} else {
			var response wire.GetNetworkInfoResponse
			err = wire.Decode(data.Bytes(), &response)
			if err != nil {
				return err
			}
			logrus.Infof("Response: %+v", response)
		}
	case 0x1107:
		logrus.Infof("Function: Add or modify permissions")
		logrus.Warnf("TODO NOT IMPLEMENTED")
	case 0x1108:
		logrus.Infof("Function: Delete an authority")
		logrus.Warnf("TODO NOT IMPLEMENTED")
	case 0x11F2:
		logrus.Infof("Function: Setting TCPIP")
		logrus.Warnf("TODO NOT IMPLEMENTED")
	default:
		logrus.Warnf("TODO UNHANDLED FUNCTION: %X", envelope.Function)
	}
	return nil
}

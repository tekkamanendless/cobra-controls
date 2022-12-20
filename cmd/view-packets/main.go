package main

import (
	"fmt"
	"os"

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
			data := packet.TransportLayer().LayerPayload()
			logrus.Infof("Data (%d): %X", len(data), data)

			parseData(data)
		}
	}
}

func parseData(data []byte) {
	if len(data) < 5 {
		logrus.Warnf("Unhandled length: %d", len(data))
		return
	}
	startByte := data[0]
	data = data[1:]
	boardAddress := []byte{data[1], data[0]}
	data = data[2:]
	functionType := []byte{data[1], data[0]}
	data = data[2:]

	logrus.Infof("Packet:")
	logrus.Infof("Start byte: %X", startByte)
	logrus.Infof("Board address: %X", boardAddress)
	logrus.Infof("Function type: %X", functionType)
	logrus.Infof("Remaining data: (%d) %X", len(data), data)

	switch fmt.Sprintf("%X", functionType) {
	case "1081":
		logrus.Infof("Function: Read Operation Status Information")
	case "108B":
		logrus.Infof("Function: Set the time")
	case "108D":
		logrus.Infof("Function: Read the records information (by index)")
	case "108E":
		logrus.Infof("Function: Remove a specified number of records")
	case "108F":
		logrus.Infof("Function: Set door controls (online/delay)")
	case "1093":
		logrus.Infof("Function: Clearr popedom")
	case "1095":
		logrus.Infof("Function: Read popedom")
	case "1096":
		logrus.Infof("Function: Read control period of time")
	case "1097":
		logrus.Infof("Function: Modification control period of time")
	case "109B":
		logrus.Infof("Function: Tail plus permissions")
	case "109D":
		logrus.Infof("Function: Long-distance open door")
	case "10F1":
		logrus.Infof("Function: Read")
	case "10F4":
		logrus.Infof("Function: Amend")
	case "1107":
		logrus.Infof("Function: Add or modify permissions")
	case "1108":
		logrus.Infof("Function: Delete an authority")
	default:
		logrus.Warnf("Unhandled function: %X", functionType)
	}
}

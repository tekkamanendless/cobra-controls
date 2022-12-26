package main

import (
	"flag"
	"fmt"

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
	items := flag.Args()

	for i, item := range items {
		var data []byte
		fmt.Sscanf(item, "%x", &data)

		logrus.Infof("Item[%2d]: %x", i, data)

		logrus.Infof("Read Uint8.")
		for d := 0; d < len(data); d++ {
			reader := wire.NewReader(data[d:])
			value, _ := reader.ReadUint8()
			logrus.Infof("   %2d: %3d (0x%2x)", d, value, value)
		}

		logrus.Infof("Read Uint16.")
		for d := 0; d < len(data); d++ {
			reader := wire.NewReader(data[d:])
			value, _ := reader.ReadUint16()
			logrus.Infof("   %2d: %5d (0x%4x)", d, value, value)
		}

		logrus.Infof("Read Uint24.")
		for d := 0; d < len(data); d++ {
			reader := wire.NewReader(data[d:])
			value, _ := reader.ReadUint24()
			logrus.Infof("   %2d: %8d (0x%6x)", d, value, value)
		}

		logrus.Infof("Read Uint32.")
		for d := 0; d < len(data); d++ {
			reader := wire.NewReader(data[d:])
			value, _ := reader.ReadUint32()
			logrus.Infof("   %2d: %11d (0x%8x)", d, value, value)
		}

		logrus.Infof("Read date.")
		for d := 0; d < len(data); d++ {
			reader := wire.NewReader(data[d:])
			value, _ := reader.ReadDate()
			logrus.Infof("   %2d: %v", d, value)
		}

		logrus.Infof("Read time.")
		for d := 0; d < len(data); d++ {
			reader := wire.NewReader(data[d:])
			value, _ := reader.ReadTime()
			logrus.Infof("   %2d: %v", d, value)
		}
	}
}

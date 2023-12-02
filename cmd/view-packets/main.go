package main

import (
	"fmt"
	"os"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gitlab.com/doug-manley/cobra-controls/cobrafile"
	"gitlab.com/doug-manley/cobra-controls/wire"
)

func main() {
	var controllerFile string
	var personnelFile string

	var controllerList cobrafile.ControllerList
	var personnelList cobrafile.PersonnelList
	verbose := false

	rootCommand := &cobra.Command{
		Use:   "view-packets <pcap-file>[ ...]",
		Short: "View packets",
		Long:  ``,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if verbose {
				logrus.SetLevel(logrus.DebugLevel)
			}

			if controllerFile != "" {
				var err error
				controllerList, err = cobrafile.LoadController(controllerFile)
				if err != nil {
					logrus.Warnf("Could not load controller file: %v", err)
				}
				logrus.Infof("Controllers: (%d)", len(controllerList))
			}

			if personnelFile != "" {
				var err error
				personnelList, err = cobrafile.LoadPersonnel(personnelFile)
				if err != nil {
					logrus.Warnf("Could not load personnel file: %v", err)
				}
				logrus.Infof("Personnel: (%d)", len(personnelList))
			}
		},
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			filenames := args

			filterPacket := func(packet gopacket.Packet) bool {
				if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
					logrus.Debugf("This is a TCP packet.")
					tcp, _ := tcpLayer.(*layers.TCP)
					logrus.Debugf("From src port %d to dst port %d.", tcp.SrcPort, tcp.DstPort)
					if tcp.SrcPort != wire.PortDefault && tcp.DstPort != wire.PortDefault {
						return false
					}
				} else if udpLayer := packet.Layer(layers.LayerTypeUDP); udpLayer != nil {
					logrus.Debugf("This is a UDP packet.")
					udp, _ := udpLayer.(*layers.UDP)
					logrus.Debugf("From src port %d to dst port %d.", udp.SrcPort, udp.DstPort)
					if udp.SrcPort != wire.PortDefault && udp.DstPort != wire.PortDefault {
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
					var controllerAddress string
					if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
						if tcp, ok := tcpLayer.(*layers.TCP); ok {
							if tcp.DstPort == wire.PortDefault {
								fromClient = true
								controllerAddress = packet.NetworkLayer().NetworkFlow().Dst().String()
							} else {
								controllerAddress = packet.NetworkLayer().NetworkFlow().Src().String()
							}
						}
					} else if udpLayer := packet.Layer(layers.LayerTypeUDP); udpLayer != nil {
						if udp, ok := udpLayer.(*layers.UDP); ok {
							if udp.DstPort == wire.PortDefault {
								fromClient = true
								controllerAddress = packet.NetworkLayer().NetworkFlow().Dst().String()
							} else {
								controllerAddress = packet.NetworkLayer().NetworkFlow().Src().String()
							}
						}
					} else {
						logrus.Warnf("Could not determine source/destination from packet.")
					}

					data := packet.TransportLayer().LayerPayload()
					logrus.Debugf("Data (%d): %X", len(data), data)

					err = parseData(wire.NewReader(data), fromClient, controllerAddress, controllerList, personnelList)
					if err != nil {
						logrus.Warnf("Could not parse data: [%T] %v", err, err)
					}
				}
			}
		},
	}
	rootCommand.PersistentFlags().StringVar(&controllerFile, "controller-file", "", "Use this CSV file to load the controller information")
	rootCommand.PersistentFlags().StringVar(&personnelFile, "personnel-file", "", "Use this CSV file to load the personnel information")
	rootCommand.PersistentFlags().BoolVar(&verbose, "verbose", false, "Enable verbose output")

	err := rootCommand.Execute()
	if err != nil {
		logrus.Errorf("Error: %v", err)
	}
	os.Exit(0)
}

func parseData(fullContents *wire.Reader, fromClient bool, controllerAddress string, controllerList cobrafile.ControllerList, personnelList cobrafile.PersonnelList) error {
	if fromClient {
		logrus.Infof("From: Client")
	} else {
		logrus.Infof("From: Controller")
	}

	var envelope wire.Envelope
	err := wire.Decode(fullContents, &envelope)
	if err != nil {
		return fmt.Errorf("could not decode envelope: %w", err)
	}

	logrus.Infof("Packet:")
	if fromClient {
		logrus.Infof("   Source: %s", controllerAddress)
	} else {
		var controllerName string
		if controllerList != nil {
			controllerName = controllerList.LookupName(controllerAddress)
		}
		logrus.Infof("   Controller: %s (name: %s)", controllerAddress, controllerName)
	}
	logrus.Infof("   Board address: 0x%X", envelope.BoardAddress)
	logrus.Infof("   Function type: 0x%X", envelope.Function)
	logrus.Infof("   Remaining data: (%d) %X", len(envelope.Contents), envelope.Contents)

	data := wire.NewReader(envelope.Contents)

	switch envelope.Function {
	case wire.FunctionGetOperationStatus:
		logrus.Infof("Function: GetOperationStatus")
		if fromClient {
			var request wire.GetOperationStatusRequest
			err = wire.Decode(data, &request)
			if err != nil {
				return err
			}
			logrus.Infof("Request: %+v", request)
		} else {
			var response wire.GetOperationStatusResponse
			err = wire.Decode(data, &response)
			if err != nil {
				return err
			}
			logrus.Infof("Response: %+v", response)
			if response.Record != nil {
				logrus.Infof("Record: %+v", *response.Record)
				if personnelList != nil {
					if person := personnelList.FindByCardID(wire.CardID(response.Record.AreaNumber, response.Record.IDNumber)); person != nil {
						logrus.Infof("   Person: %+v", *person)
					}
				}
				if response.Record.AreaNumber == 0 && response.Record.IDNumber < 100 {
					if response.Record.RecordState == 0b00 && response.Record.IDNumber <= 3 {
						logrus.Infof("   Action: Button")
					} else if response.Record.RecordState == 0b11 && response.Record.IDNumber <= 3 {
						logrus.Infof("   Action: Remote control")
					} else {
						logrus.Warnf("   Action: TODO UNHANDLED")
					}
				}
				logrus.Infof("   Door: %d (access: %t)", response.Record.Door(), response.Record.AccessGranted())
				if controllerList != nil {
					door := controllerList.LookupDoor(controllerAddress, response.Record.Door())
					if door != "" {
						logrus.Infof("   Door name: %s", door)
					}
				}
			}
		}
	case wire.FunctionGetBasicInfo:
		logrus.Infof("Function: GetBasicInfo")
		if fromClient {
			var request wire.GetBasicInfoRequest
			err = wire.Decode(data, &request)
			if err != nil {
				return err
			}
			logrus.Infof("Request: %+v", request)
		} else {
			var response wire.GetBasicInfoResponse
			err = wire.Decode(data, &response)
			if err != nil {
				return err
			}
			logrus.Infof("Response: %+v", response)
		}
	case wire.FunctionSetTime:
		logrus.Infof("Function: SetTime")
		if fromClient {
			var request wire.SetTimeRequest
			err = wire.Decode(data, &request)
			if err != nil {
				return err
			}
			logrus.Infof("Request: %+v", request)
		} else {
			var response wire.SetTimeResponse
			err = wire.Decode(data, &response)
			if err != nil {
				return err
			}
			logrus.Infof("Response: %+v", response)
		}
	case wire.FunctionGetRecord:
		logrus.Infof("Function: GetRecord")
		if fromClient {
			var request wire.GetRecordRequest
			err = wire.Decode(data, &request)
			if err != nil {
				return err
			}
			logrus.Infof("Request: %+v", request)
		} else {
			var response wire.GetRecordResponse
			err = wire.Decode(data, &response)
			if err != nil {
				return err
			}
			logrus.Infof("Response: %+v", response)
		}
	case wire.FunctionDeleteRecord:
		logrus.Infof("Function: DeleteRecord")
		if fromClient {
			var request wire.DeleteRecordRequest
			err = wire.Decode(data, &request)
			if err != nil {
				return err
			}
			logrus.Infof("Request: %+v", request)
		} else {
			var response wire.DeleteRecordResponse
			err = wire.Decode(data, &response)
			if err != nil {
				return err
			}
			logrus.Infof("Response: %+v", response)
		}
	case 0x108F:
		logrus.Infof("Function: Set door controls (online/delay)")
		logrus.Warnf("TODO NOT IMPLEMENTED")
	case 0x1091:
		logrus.Infof("Function: Upload the mission")
		logrus.Warnf("TODO NOT IMPLEMENTED")
	case wire.FunctionClearUpload:
		logrus.Infof("Function: ClearUpload")
		if fromClient {
			var request wire.ClearUploadRequest
			err = wire.Decode(data, &request)
			if err != nil {
				return err
			}
			logrus.Infof("Request: %+v", request)
		} else {
			var response wire.ClearUploadResponse
			err = wire.Decode(data, &response)
			if err != nil {
				return err
			}
			logrus.Infof("Response: %+v", response)
		}
	case 0x1096:
		logrus.Infof("Function: Read control period of time")
		logrus.Warnf("TODO NOT IMPLEMENTED")
	case wire.FunctionUpdateControlPeriod:
		logrus.Infof("Function: UpdateControlPeriod")
		if fromClient {
			var request wire.UpdateControlPeriodRequest
			err = wire.Decode(data, &request)
			if err != nil {
				return err
			}
			logrus.Infof("Request: %+v", request)
		} else {
			var response wire.UpdateControlPeriodResponse
			err = wire.Decode(data, &response)
			if err != nil {
				return err
			}
			logrus.Infof("Response: %+v", response)
		}

	case wire.FunctionUnknown1098:
		logrus.Infof("Function: Unknown1098")
		if fromClient {
			var request wire.Unknown1098Request
			err = wire.Decode(data, &request)
			if err != nil {
				return err
			}
			logrus.Infof("Request: %+v", request)
		} else {
			var response wire.Unknown1098Response
			err = wire.Decode(data, &response)
			if err != nil {
				return err
			}
			logrus.Infof("Response: %+v", response)
		}
	case wire.FunctionTailPlusPermissions:
		logrus.Infof("Function: FunctionTailPlusPermissions")
		if fromClient {
			var request wire.TailPlusPermissionsRequest
			err = wire.Decode(data, &request)
			if err != nil {
				return err
			}
			logrus.Infof("Request: %+v", request)
			if personnelList != nil {
				if person := personnelList.FindByCardID(wire.CardID(request.AreaNumber, request.CardNumber)); person != nil {
					logrus.Infof("   Person: %+v", *person)
				}
			}
			if controllerList != nil {
				door := controllerList.LookupDoor(controllerAddress, request.Door)
				if door != "" {
					logrus.Infof("Door: %s", door)
				}
			}
		} else {
			var response wire.TailPlusPermissionsResponse
			err = wire.Decode(data, &response)
			if err != nil {
				return err
			}
			logrus.Infof("Response: %+v", response)
		}
	case wire.FunctionOpenDoor:
		logrus.Infof("Function: OpenDoor")
		if fromClient {
			var request wire.OpenDoorRequest
			err = wire.Decode(data, &request)
			if err != nil {
				return err
			}
			logrus.Infof("Request: %+v", request)
			if controllerList != nil {
				door := controllerList.LookupDoor(controllerAddress, request.Door)
				if door != "" {
					logrus.Infof("Door: %s", door)
				}
			}
		} else {
			var response wire.OpenDoorResponse
			err = wire.Decode(data, &response)
			if err != nil {
				return err
			}
			logrus.Infof("Response: %+v", response)
		}
	case wire.FunctionGetSetting:
		logrus.Infof("Function: GetSetting")
		if fromClient {
			var request wire.GetSettingRequest
			err = wire.Decode(data, &request)
			if err != nil {
				return err
			}
			logrus.Infof("Request: %+v", request)
		} else {
			var response wire.GetSettingResponse
			err = wire.Decode(data, &response)
			if err != nil {
				return err
			}
			logrus.Infof("Response: %+v", response)
		}
	case wire.FunctionUpdateSetting:
		logrus.Infof("Function: UpdateSetting")
		if fromClient {
			var request wire.UpdateSettingRequest
			err = wire.Decode(data, &request)
			if err != nil {
				return err
			}
			logrus.Infof("Request: %+v", request)
		} else {
			var response wire.UpdateSettingResponse
			err = wire.Decode(data, &response)
			if err != nil {
				return err
			}
			logrus.Infof("Response: %+v", response)
		}
	case wire.FunctionGetUpload:
		logrus.Infof("Function: GetUpload")
		if fromClient {
			var request wire.GetUploadRequest
			err = wire.Decode(data, &request)
			if err != nil {
				return err
			}
			logrus.Infof("Request: %+v", request)
		} else {
			var response wire.GetUploadResponse
			err = wire.Decode(data, &response)
			if err != nil {
				return err
			}
			logrus.Infof("Response: %+v", response)
			if personnelList != nil {
				if person := personnelList.FindByCardID(wire.CardID(response.AreaNumber, response.IDNumber)); person != nil {
					logrus.Infof("   Person: %+v", *person)
				}
			}
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
			logrus.Infof("Unknown1: %d", unknown1)                             // This seems to always be "3".
			logrus.Infof("Unknown2: %d", unknown2)                             // This seems to always be "0".
			logrus.Infof("Unknown3: %d (index, maybe?)", unknown3)             // This seems to be the 0-index for access uploads.
			logrus.Infof("Unknown4: %d (1 for basic, 4 for access)", unknown4) // This seems to be 1 for basic upload, 4 for access upload.
			// TODO: The remaining data appears to be different based on Unknown4.
			// TODO: For "basic" uploads, I don't know what this is yet.
			// TODO: For "access" uploads (Unknown4=4), this is a list of popedom records.
			switch unknown4 {
			case 1:
				newData, err := data.Read(270)
				if err != nil {
					return fmt.Errorf("could not read proper payload: %w", err)
				}
				if !wire.IsAll(data.Bytes(), 0x00) {
					logrus.Infof("Remainder is not all 0x00: %x", data.Bytes())
				}

				data = newData
				switch unknown3 {
				case 1:
					unknown5, err := data.ReadUint16()
					if err != nil {
						return fmt.Errorf("could not read unknown5: %w", err)
					}
					logrus.Infof("Unknown5: %d", unknown5)

					openDelay1, err := data.ReadUint16()
					if err != nil {
						return fmt.Errorf("could not read openDelay1: %w", err)
					}
					logrus.Infof("Open delay 1: %d seconds", openDelay1/10)
					openDelay2, err := data.ReadUint16()
					if err != nil {
						return fmt.Errorf("could not read openDelay2: %w", err)
					}
					logrus.Infof("Open delay 2: %d seconds", openDelay2/10)
					openDelay3, err := data.ReadUint16()
					if err != nil {
						return fmt.Errorf("could not read openDelay3: %w", err)
					}
					logrus.Infof("Open delay 3: %d seconds", openDelay3/10)
					openDelay4, err := data.ReadUint16()
					if err != nil {
						return fmt.Errorf("could not read openDelay4: %w", err)
					}
					logrus.Infof("Open delay 4: %d seconds", openDelay4/10)

					controlState1, err := data.ReadUint8()
					if err != nil {
						return fmt.Errorf("could not read controlState1: %w", err)
					}
					logrus.Infof("Control state 1: %d (1 is open, 2 is closed, 3 is door controlled)", controlState1)
					controlState2, err := data.ReadUint8()
					if err != nil {
						return fmt.Errorf("could not read controlState2: %w", err)
					}
					logrus.Infof("Control state 2: %d (1 is open, 2 is closed, 3 is door controlled)", controlState2)
					controlState3, err := data.ReadUint8()
					if err != nil {
						return fmt.Errorf("could not read controlState3: %w", err)
					}
					logrus.Infof("Control state 3: %d (1 is open, 2 is closed, 3 is door controlled)", controlState3)
					controlState4, err := data.ReadUint8()
					if err != nil {
						return fmt.Errorf("could not read controlState4: %w", err)
					}
					logrus.Infof("Control state 4: %d (1 is open, 2 is closed, 3 is door controlled)", controlState4)

					// Remainder example:
					// 000000000000000000000000000000010100100000fa006401015500000000700000000000000000000000000000000000000000000000000000000084941309ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000d00000000000000000000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000ff000000000000000000000000000000000000000000000000000000000000000000fcffc3e3f929001000fcffffff0f
					//                                     ^|                                                                                      ^                                                              |
					//                                     Invalid card swiping                                                                    \______________________________________________________________/
					//                                                                                                                               Passwords (list of 16-bit numbers)

					remainder, err := data.Read(data.Length())
					if err != nil {
						return fmt.Errorf("could not read data: %w", err)
					}
					logrus.Infof("Unknown10F9_%d: %x", unknown4, remainder.Bytes())
				default:
					if wire.IsAll(data.Bytes(), 0xff) {
						logrus.Infof("Remainder is all 0xff.")
					} else {
						remainder, err := data.Read(data.Length())
						if err != nil {
							return fmt.Errorf("could not read data: %w", err)
						}
						logrus.Infof("Unknown10F9_%d_%d: %x", unknown4, unknown3, remainder.Bytes())
					}
				}
			case 4:
				for p := 0; data.Length() >= 16; p++ {
					popedom, err := data.Read(16)
					if err != nil {
						logrus.Warnf("could not read popedom [%d]: %v", p, err)
						continue
					}
					logrus.Infof("Popedom[%2d]: %X", p, popedom.Bytes())
					if wire.IsAll(popedom.Bytes(), 0xff) {
						logrus.Infof("Skipping bogus popedom.")
						continue
					}
					err = func(popedom *wire.Reader) error {
						id, err := popedom.ReadUint16()
						if err != nil {
							return fmt.Errorf("could not read id: %w", err)
						}
						area, err := popedom.ReadUint8()
						if err != nil {
							return fmt.Errorf("could not read area: %w", err)
						}
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
						logrus.Infof("Card ID: %s", wire.CardID(area, id))
						if personnelList != nil {
							if person := personnelList.FindByCardID(wire.CardID(area, id)); person != nil {
								logrus.Infof("   Person: %+v", *person)
							}
						}
						logrus.Infof("Door: %d", door)
						logrus.Infof("Open Date: %v", openDate)
						logrus.Infof("Close Date: %v", closeDate)
						logrus.Infof("Control index: %X", controlIndex) // 0 to not use control time; 1 to specify a time.
						logrus.Infof("Password: %X", password)
						logrus.Infof("Standby 1: %X", standby1) // 1 for the "first card users"; 0 for not those users.
						logrus.Infof("Standby 2: %X", standby2) // 0 for the general user group; >0 for special group permissions.
						logrus.Infof("Standby 3: %X", standby3)
						logrus.Infof("Standby 4: %X", standby4)

						return nil
					}(popedom)
					if err != nil {
						logrus.Warnf("could not parse popedom [%d]: %v", p, err)
						continue
					}
				}
				if data.Length() != 0 {
					logrus.Warnf("Unexpected trailing data length: (%d)", data.Length())
				}
			default:
				remainder, err := data.Read(data.Length())
				if err != nil {
					return fmt.Errorf("could not read data: %w", err)
				}
				logrus.Infof("Unknown10F9_%d: %x", unknown4, remainder.Bytes())
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
			err = wire.Decode(data, &request)
			if err != nil {
				return err
			}
			logrus.Infof("Request: %+v", request)
		} else {
			var response wire.GetNetworkInfoResponse
			err = wire.Decode(data, &response)
			if err != nil {
				return err
			}
			logrus.Infof("Response: %+v", response)
		}
	case wire.FunctionUpdatePermissions:
		logrus.Infof("Function: UpdatePermissions")
		if fromClient {
			var request wire.UpdatePermissionsRequest
			err = wire.Decode(data, &request)
			if err != nil {
				return err
			}
			logrus.Infof("Request: %+v", request)
			logrus.Infof("Card ID: %s", wire.CardID(request.Area, request.CardID))
			if personnelList != nil {
				if person := personnelList.FindByCardID(wire.CardID(request.Area, request.CardID)); person != nil {
					logrus.Infof("   Person: %+v", *person)
				}
			}
		} else {
			var response wire.UpdatePermissionsResponse
			err = wire.Decode(data, &response)
			if err != nil {
				return err
			}
			logrus.Infof("Response: %+v", response)
		}
	case wire.FunctionDeletePermissions:
		logrus.Infof("Function: DeletePermissions")
		if fromClient {
			var request wire.DeletePermissionsRequest
			err = wire.Decode(data, &request)
			if err != nil {
				return err
			}
			logrus.Infof("Request: %+v", request)
			logrus.Infof("Card ID: %s", wire.CardID(request.Area, request.CardID))
			if personnelList != nil {
				if person := personnelList.FindByCardID(wire.CardID(request.Area, request.CardID)); person != nil {
					logrus.Infof("   Person: %+v", *person)
				}
			}
		} else {
			var response wire.DeletePermissionsResponse
			err = wire.Decode(data, &response)
			if err != nil {
				return err
			}
			logrus.Infof("Response: %+v", response)
		}
	case wire.FunctionSetNetworkInfo:
		logrus.Infof("Function: SetNetworkInfo")
		if fromClient {
			var request wire.SetNetworkInfoRequest
			err = wire.Decode(data, &request)
			if err != nil {
				return err
			}
			logrus.Infof("Request: %+v", request)
		} else {
			var response wire.SetNetworkInfoResponse
			err = wire.Decode(data, &response)
			if err != nil {
				return err
			}
			logrus.Infof("Response: %+v", response)
		}

	default:
		logrus.Warnf("TODO UNHANDLED FUNCTION: %X", envelope.Function)
	}
	return nil
}

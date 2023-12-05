package main

import (
	"fmt"
	"net"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tekkamanendless/cobra-controls/cobrafile"
	"github.com/tekkamanendless/cobra-controls/wire"
)

func main() {
	var controllerName string
	var controllerAddress string
	var controllerPort uint16
	var boardAddressString string
	var boardAddress uint16
	var controllerFile string
	var personnelFile string
	var protocol string

	var clients []*wire.Client
	var controllerList cobrafile.ControllerList
	var personnelList cobrafile.PersonnelList
	verbose := false

	rootCommand := &cobra.Command{
		Use:   "cobra-cli",
		Short: "Command-line tools for Cobra Controls access systems",
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
				logrus.Debugf("Controllers: (%d)", len(controllerList))
			}

			if personnelFile != "" {
				var err error
				personnelList, err = cobrafile.LoadPersonnel(personnelFile)
				if err != nil {
					logrus.Warnf("Could not load personnel file: %v", err)
				}
				logrus.Debugf("Personnel: (%d)", len(personnelList))
			}

			if len(boardAddressString) > 0 {
				v, err := strconv.ParseInt(boardAddressString, 0 /*auto-detect base*/, 17 /*one more than 16 because this is signed*/)
				if err != nil {
					logrus.Errorf("Could not parse board address: %v", err)
					os.Exit(1)
				}
				boardAddress = uint16(v)
				logrus.Debugf("Board address: %d (0x%x)", boardAddress, boardAddress)
			}

			if controllerAddress != "" && controllerPort > 0 && boardAddress > 0 {
				client := &wire.Client{
					ControllerAddress: controllerAddress,
					ControllerPort:    uint16(controllerPort),
					BoardAddress:      boardAddress,
					Protocol:          wire.Protocol(protocol),
				}
				logrus.Debugf("Client: %+v", client)
				clients = append(clients, client)
			} else if controllerName != "" {
				_, err := path.Match(controllerName, "PLACEHOLDER TEXT")
				if err != nil {
					logrus.Errorf("Invalid matching expression: %v", err)
					os.Exit(1)
				}
				for _, controller := range controllerList {
					if ok, _ := path.Match(controllerName, controller.Name); ok {
						client := &wire.Client{
							ControllerAddress: controller.Address,
							ControllerPort:    controller.Port,
							BoardAddress:      controller.SN,
							Protocol:          wire.Protocol(protocol),
						}
						logrus.Debugf("Client: %+v", client)
						clients = append(clients, client)
					}
				}
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
			os.Exit(1)
		},
	}
	rootCommand.PersistentFlags().StringVar(&controllerName, "controller-name", "", "A wildcard expression to match controllers")
	rootCommand.PersistentFlags().StringVar(&controllerAddress, "controller-address", "", "Set the controller address")
	rootCommand.PersistentFlags().Uint16Var(&controllerPort, "controller-port", wire.PortDefault, "Set the controller address")
	rootCommand.PersistentFlags().StringVar(&boardAddressString, "board-address", "", "Set the board address (either hexadecimal or decimial)")
	rootCommand.PersistentFlags().StringVar(&controllerFile, "controller-file", "", "Use this CSV file to load the controller information")
	rootCommand.PersistentFlags().StringVar(&personnelFile, "personnel-file", "", "Use this CSV file to load the personnel information")
	rootCommand.PersistentFlags().StringVar(&protocol, "protocol", "", "Use this protocol to communicate (if unspecified, the appropriate default for the command will be used)")
	rootCommand.PersistentFlags().BoolVar(&verbose, "verbose", false, "Enable verbose output")

	{
		cmd := &cobra.Command{
			Use:   "drift",
			Short: "Show the time drift",
			Long:  ``,
			Args:  cobra.NoArgs,
			Run: func(cmd *cobra.Command, args []string) {
				if len(clients) == 0 {
					logrus.Errorf("Invalid client")
					os.Exit(1)
				}

				for _, client := range clients {
					var controller string
					if controllerList != nil {
						controller = controllerList.LookupName(client.ControllerAddress)
					}
					if controller == "" {
						controller = client.ControllerAddress
					}

					var sum time.Duration
					count := 0
					for i := 0; i < 10; i++ {
						currentTime := time.Now()
						currentTime = time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), currentTime.Hour(), currentTime.Minute(), currentTime.Second(), 0, time.UTC)

						request := wire.GetOperationStatusRequest{
							RecordIndex: 0,
						}
						var response wire.GetOperationStatusResponse
						err := client.Do(wire.FunctionGetOperationStatus, &request, &response)
						if err != nil {
							logrus.Errorf("Error from client: %v", err)
							continue
						}

						logrus.Debugf("Current time: %v", currentTime)
						logrus.Debugf("System time: %v", response.CurrentTime)
						timeAhead := response.CurrentTime.Sub(currentTime)
						sum += timeAhead
						count++
					}
					drift := sum / time.Duration(count)
					if count == 0 {
						fmt.Printf("Controller: %s | Drift: unknown\n", controller)
					} else {
						fmt.Printf("Controller: %s | Drift: %v (+ is ahead, - is behind)\n", controller, drift)
					}
				}
			},
		}

		rootCommand.AddCommand(cmd)
	}

	{
		cmd := &cobra.Command{
			Use:   "get-upload <index>[ ...]",
			Short: "Get an upload record",
			Long:  ``,
			Args:  cobra.MinimumNArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				if len(clients) == 0 {
					logrus.Errorf("Invalid client")
					os.Exit(1)
				}

				var indexes []uint16
				for _, arg := range args {
					v, err := strconv.ParseInt(arg, 0, 17)
					if err != nil {
						logrus.Errorf("Could not parse index: %v", err)
						os.Exit(1)
					}
					indexes = append(indexes, uint16(v))
				}

				for _, client := range clients {
					for _, index := range indexes {
						logrus.Debugf("Index: %d", index)
						request := wire.GetUploadRequest{
							Index: index,
						}
						var response wire.GetUploadResponse
						err := client.Do(wire.FunctionGetUpload, &request, &response)
						if err != nil {
							logrus.Errorf("Error from client: %v", err)
							continue
						}
						logrus.Debugf("Response: %+v", response)

						var controller string
						var door string
						var person *cobrafile.Person
						if controllerList != nil {
							controller, door = controllerList.LookupNameAndDoor(client.ControllerAddress, response.DoorNumber)
						}
						if personnelList != nil {
							person = personnelList.FindByCardID(wire.CardID(response.AreaNumber, response.IDNumber))
						}
						if controller == "" {
							controller = client.ControllerAddress
						}
						if door == "" {
							door = fmt.Sprintf("%d", response.DoorNumber)
						}
						if person == nil {
							fmt.Printf("Controller: %s | Index: %d | Door: %s | Card ID: %s | Response: %+v\n", controller, request.Index, door, wire.CardID(response.AreaNumber, response.IDNumber), response)
						} else {
							fmt.Printf("Controller: %s | Index: %d | Door: %s | Card ID: %s | Name: %s | Response: %+v\n", controller, request.Index, door, wire.CardID(response.AreaNumber, response.IDNumber), person.Name, response)
						}

					}
				}
			},
		}

		rootCommand.AddCommand(cmd)
	}

	{
		cmd := &cobra.Command{
			Use:   "history [<index|range>[ ...]]",
			Short: "Query the history",
			Long:  `You may specify either a specific index, such as "1234", or a range of the form "last #", such as "last 50".  If nothing is specified, then the most recent record is shown.`,
			Args:  cobra.MinimumNArgs(0),
			Run: func(cmd *cobra.Command, args []string) {
				if len(clients) == 0 {
					logrus.Errorf("Invalid client")
					os.Exit(1)
				}

				nextNumbersByClient := make([][]uint32, len(clients))
				for i := 0; i < len(args); i++ {
					arg := args[i]
					switch arg {
					case "last":
						i++
						if i >= len(args) {
							logrus.Errorf("Expected number after 'last'.")
							os.Exit(1)
						}
						arg = args[i]
						v, err := strconv.ParseInt(arg, 0, 33)
						if err != nil {
							logrus.Errorf("Could not parse value %q: %v", arg, err)
							os.Exit(1)
						}
						for c, client := range clients {
							request := wire.GetOperationStatusRequest{
								RecordIndex: 0,
							}
							var response wire.GetOperationStatusResponse
							err := client.Do(wire.FunctionGetOperationStatus, &request, &response)
							if err != nil {
								logrus.Errorf("Error from client: %v", err)
								continue
							}
							logrus.Debugf("Response: %+v", response)

							startIndex := uint32(1)
							if response.RecordCount > uint32(v) {
								startIndex = response.RecordCount - uint32(v)
							}
							for i := startIndex; i <= response.RecordCount; i++ {
								nextNumbersByClient[c] = append(nextNumbersByClient[c], i)
							}
						}
					default:
						v, err := strconv.ParseInt(arg, 0, 33)
						if err != nil {
							logrus.Errorf("Could not parse value %q: %v", arg, err)
							os.Exit(1)
						}
						for c := range clients {
							nextNumbersByClient[c] = []uint32{uint32(v)}
						}
					}
				}

				for c, client := range clients {
					nextNumbers := nextNumbersByClient[c]
					if len(nextNumbers) == 0 {
						nextNumbers = append(nextNumbers, 0)
					}
					for _, nextNumber := range nextNumbers {
						logrus.Debugf("Next number: %d", nextNumber)
						request := wire.GetOperationStatusRequest{
							RecordIndex: nextNumber,
						}
						var response wire.GetOperationStatusResponse
						err := client.Do(wire.FunctionGetOperationStatus, &request, &response)
						if err != nil {
							logrus.Errorf("Error from client: %v", err)
							continue
						}
						logrus.Debugf("Response: %+v", response)
						if response.Record != nil {
							logrus.Debugf("Record: %+v", *response.Record)
							var person *cobrafile.Person
							var controller, door string
							if controllerList != nil {
								controller, door = controllerList.LookupNameAndDoor(client.ControllerAddress, response.Record.Door())
							}
							if personnelList != nil {
								person = personnelList.FindByCardID(wire.CardID(response.Record.AreaNumber, response.Record.IDNumber))
							}
							if controller == "" {
								controller = client.ControllerAddress
							}
							if door == "" {
								door = fmt.Sprintf("%d", response.Record.RecordState)
							}
							if person == nil {
								fmt.Printf("%v | Controller: %s | Index: %d | Door: %s | Card ID: %s | Access: %t\n", response.Record.BrushDateTime, controller, request.RecordIndex, door, wire.CardID(response.Record.AreaNumber, response.Record.IDNumber), response.Record.AccessGranted())
							} else {
								fmt.Printf("%v | Controller: %s | Index: %d | Door: %s | Card ID: %s | Name: %s | Access: %t\n", response.Record.BrushDateTime, controller, request.RecordIndex, door, wire.CardID(response.Record.AreaNumber, response.Record.IDNumber), person.Name, response.Record.AccessGranted())
							}
						}
					}
				}
			},
		}

		rootCommand.AddCommand(cmd)
	}

	{
		cmd := &cobra.Command{
			Use:   "info",
			Short: "Gather information",
			Long:  ``,
			Args:  cobra.NoArgs,
			Run: func(cmd *cobra.Command, args []string) {
				if len(clients) == 0 {
					logrus.Errorf("Invalid client")
					os.Exit(1)
				}

				for _, client := range clients {
					{
						var response wire.GetBasicInfoResponse
						err := client.Do(wire.FunctionGetBasicInfo, nil, &response)
						if err != nil {
							logrus.Errorf("Error: %v", err)
							continue
						}
						logrus.Infof("Response: %+v", response)
					}
					{
						request := &wire.GetNetworkInfoRequest{
							Unknown1: 1,
						}
						var response wire.GetNetworkInfoResponse
						err := client.Do(wire.FunctionGetNetworkInfo, request, &response)
						if err != nil {
							logrus.Errorf("Error: %v", err)
							continue
						}
						logrus.Infof("Response: %+v", response)
					}
				}
			},
		}

		rootCommand.AddCommand(cmd)
	}

	{
		var batchCount int
		var sleepDuration time.Duration

		cmd := &cobra.Command{
			Use:   "monitor",
			Short: "Monitor a door",
			Long:  ``,
			Args:  cobra.NoArgs,
			Run: func(cmd *cobra.Command, args []string) {
				if len(clients) == 0 {
					logrus.Errorf("Invalid client")
					os.Exit(1)
				}

				nextNumbers := make([]uint32, len(clients)) // If this is zero, then we'll ask for the latest value.
				for batch := 0; ; batch++ {
					if batchCount > 0 {
						if batch >= batchCount {
							break
						}
					}
					if batch > 0 {
						logrus.Debugf("Sleeping for %v.", sleepDuration)
						time.Sleep(sleepDuration)
					}

					for clientIndex, client := range clients {
						nextNumber := nextNumbers[clientIndex]

						logrus.Debugf("Next number: %d", nextNumber)
						request := wire.GetOperationStatusRequest{
							RecordIndex: nextNumber,
						}
						var response wire.GetOperationStatusResponse
						err := client.Do(wire.FunctionGetOperationStatus, &request, &response)
						if err != nil {
							logrus.Errorf("Error from client: %v", err)
							continue
						}
						logrus.Debugf("Response: %+v", response)
						if response.RecordCount >= nextNumber {
							if nextNumber > 0 && response.RecordCount >= nextNumber {
								for index := nextNumber; index <= response.RecordCount; index++ {
									logrus.Debugf("Geting record %d", index)
									request := wire.GetOperationStatusRequest{
										RecordIndex: index,
									}
									var response wire.GetOperationStatusResponse
									err := client.Do(wire.FunctionGetOperationStatus, &request, &response)
									if err != nil {
										logrus.Errorf("Error from client: %v", err)
										continue
									}
									logrus.Debugf("Response: %+v", response)
									if response.Record != nil {
										logrus.Debugf("Record: %+v", *response.Record)
										var person *cobrafile.Person
										var controller, door string
										if controllerList != nil {
											controller, door = controllerList.LookupNameAndDoor(client.ControllerAddress, response.Record.Door())
										}
										if personnelList != nil {
											person = personnelList.FindByCardID(wire.CardID(response.Record.AreaNumber, response.Record.IDNumber))
										}
										if controller == "" {
											controller = client.ControllerAddress
										}
										if door == "" {
											door = fmt.Sprintf("%d", response.Record.Door())
										}
										if person == nil {
											fmt.Printf("%v | Controller: %s | Door: %s | Card ID: %s | Access: %t\n", response.Record.BrushDateTime, controller, door, wire.CardID(response.Record.AreaNumber, response.Record.IDNumber), response.Record.AccessGranted())
										} else {
											fmt.Printf("%v | Controller: %s | Door: %s | Card ID: %s | Name: %s | Access: %t\n", response.Record.BrushDateTime, controller, door, wire.CardID(response.Record.AreaNumber, response.Record.IDNumber), person.Name, response.Record.AccessGranted())
										}
									}
								}
							}
							nextNumber = response.RecordCount + 1
							nextNumbers[clientIndex] = nextNumber
						}
					}
				}
			},
		}
		cmd.Flags().IntVar(&batchCount, "batch", 10, "How many iterations to run (use 0 for infinite)")
		cmd.Flags().DurationVar(&sleepDuration, "batch-interval", 5*time.Second, "How long to wait between batches")

		rootCommand.AddCommand(cmd)
	}

	{
		cmd := &cobra.Command{
			Use:   "open-door",
			Short: "Open a door",
			Long:  ``,
			Args:  cobra.NoArgs,
			Run: func(cmd *cobra.Command, args []string) {
				if len(clients) == 0 {
					logrus.Errorf("Invalid client")
					os.Exit(1)
				}

				for _, client := range clients {
					for _, arg := range args {
						logrus.Infof("Door string: %s", arg)
						var door uint8
						switch arg {
						case "1":
							door = 1
						case "2":
							door = 2
						case "3":
							door = 3
						case "4":
							door = 4
						default:
							if controllerList != nil {
								var ok bool
								door, ok = controllerList.FindDoor(client.ControllerAddress, arg)
								if !ok {
									logrus.Warnf("No such door %q for controller %s", arg, client.ControllerAddress)
									continue
								}
							}
						}
						logrus.Infof("Door value: %d", door)
						request := wire.OpenDoorRequest{
							Door:     door,
							Unkonwn1: 1,
						}
						var response wire.OpenDoorResponse
						err := client.Do(wire.FunctionGetBasicInfo, &request, &response)
						if err != nil {
							logrus.Errorf("Error: %v", err)
							continue
						}
						logrus.Infof("Response: %+v", response)
					}
				}
			},
		}

		rootCommand.AddCommand(cmd)
	}

	{
		cmd := &cobra.Command{
			Use:   "search",
			Short: "Search for controllers on the local network",
			Long:  ``,
			Args:  cobra.NoArgs,
			Run: func(cmd *cobra.Command, args []string) {
				if len(clients) == 0 {
					client := &wire.Client{
						ControllerAddress: controllerAddress,
						BoardAddress:      boardAddress,
						ControllerPort:    controllerPort,
					}
					clients = append(clients, client)
				}
				client := clients[0]

				if len(client.Protocol) == 0 {
					client.Protocol = wire.ProtocolUDP
				}
				if client.ControllerAddress == "" {
					client.ControllerAddress = "255.255.255.255"
				}
				if client.BoardAddress == 0 {
					client.BoardAddress = 0xffff
				}
				logrus.Debugf("Client: %+v", client)

				request := wire.GetNetworkInfoRequest{
					Unknown1: 0,
				}
				var responses []wire.GetNetworkInfoResponse
				responseEnvelopes, err := client.DoWithEnvelopes(wire.FunctionGetNetworkInfo, &request, &responses)
				if err != nil {
					logrus.Errorf("Error from client: %v", err)
					return
				}

				logrus.Debugf("Responses: %+v", responses)
				for i, response := range responses {
					responseEnvelope := responseEnvelopes[i]
					logrus.Debugf("Response envelope: %+v", responseEnvelope)
					logrus.Debugf("Response: %+v", response)
					fmt.Printf("Board address: %d :: MAC address: %s (%s / %s via %s on port %d)\n", responseEnvelope.BoardAddress, response.MACAddress, response.IPAddress, response.Netmask, response.Gateway, response.Port)
				}
			},
		}

		rootCommand.AddCommand(cmd)
	}

	{
		cmd := &cobra.Command{
			Use:   "set-network <mac> <ip> <netmask> <gateway> [<port>]",
			Short: "Set the network information",
			Long:  ``,
			Args:  cobra.RangeArgs(4, 5),
			Run: func(cmd *cobra.Command, args []string) {
				newMACAddress, err := net.ParseMAC(args[0])
				if err != nil {
					logrus.Errorf("Could not parse MAC address: [%T] %v", err, err)
					return
				}
				newIPAddress := net.ParseIP(args[1])
				newNetmask := net.ParseIP(args[2])
				newGateway := net.ParseIP(args[3])
				newPort := wire.PortDefault
				if len(args) > 4 {
					v, err := strconv.ParseUint(args[4], 10, 16)
					if err != nil {
						logrus.Errorf("Could not parse port number: [%T] %v", err, err)
						return
					}
					newPort = uint16(v)
				}

				if len(clients) == 0 {
					client := &wire.Client{
						ControllerAddress: controllerAddress,
						BoardAddress:      boardAddress,
						ControllerPort:    controllerPort,
					}
					clients = append(clients, client)
				}
				client := clients[0]

				if len(client.Protocol) == 0 {
					client.Protocol = wire.ProtocolUDP
				}
				if client.ControllerAddress == "" {
					client.ControllerAddress = "255.255.255.255"
				}
				if client.BoardAddress == 0 {
					client.BoardAddress = 0xffff
				}
				logrus.Debugf("Client: %+v", client)

				request := wire.SetNetworkInfoRequest{
					MACAddress: newMACAddress,
					IPAddress:  newIPAddress,
					Netmask:    newNetmask,
					Gateway:    newGateway,
					Port:       newPort,
				}
				var response wire.SetNetworkInfoResponse
				err = client.Do(wire.FunctionSetNetworkInfo, &request, &response)
				if err != nil {
					logrus.Errorf("Error: %v", err)
					return
				}
				logrus.Debugf("Response: %+v", response)

				// TODO: Print something about the response.
			},
		}

		rootCommand.AddCommand(cmd)
	}

	{
		cmd := &cobra.Command{
			Use:   "set-time",
			Short: "Set the time",
			Long:  ``,
			Args:  cobra.NoArgs,
			Run: func(cmd *cobra.Command, args []string) {
				if len(clients) == 0 {
					logrus.Errorf("Invalid client")
					os.Exit(1)
				}

				for _, client := range clients {
					var controller string
					if controllerList != nil {
						controller = controllerList.LookupName(client.ControllerAddress)
					}
					if controller == "" {
						controller = client.ControllerAddress
					}

					currentTime := time.Now()
					currentTime = time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), currentTime.Hour(), currentTime.Minute(), currentTime.Second(), 0, time.UTC)

					request := wire.SetTimeRequest{
						CurrentTime: currentTime,
					}
					var response wire.SetTimeResponse
					err := client.Do(wire.FunctionSetTime, &request, &response)
					if err != nil {
						logrus.Errorf("Error: %v", err)
						continue
					}
					logrus.Debugf("Response: %+v", response)

					fmt.Printf("Controller: %s | Current time: %v | System time: %s\n", controller, currentTime, response.CurrentTime)
				}
			},
		}

		rootCommand.AddCommand(cmd)
	}

	err := rootCommand.Execute()
	if err != nil {
		logrus.Errorf("Error: %v", err)
	}
	os.Exit(0)
}

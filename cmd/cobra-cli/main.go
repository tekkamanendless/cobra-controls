package main

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gitlab.com/doug-manley/cobra-controls/cobrafile"
	"gitlab.com/doug-manley/cobra-controls/wire"
)

func main() {
	var controllerName string
	var controllerAddress string
	var controllerPort int
	var boardAddressString string
	var boardAddress uint16
	var controllerFile string
	var personnelFile string

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
					ControllerPort:    controllerPort,
					BoardAddress:      boardAddress,
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
	rootCommand.PersistentFlags().IntVar(&controllerPort, "controller-port", 60000, "Set the controller address")
	rootCommand.PersistentFlags().StringVar(&boardAddressString, "board-address", "", "Set the board address (either hexadecimal or decimial)")
	rootCommand.PersistentFlags().StringVar(&controllerFile, "controller-file", "", "Use this CSV file to load the controller information")
	rootCommand.PersistentFlags().StringVar(&personnelFile, "personnel-file", "", "Use this CSV file to load the personnel information")
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
						err := client.Raw(wire.FunctionGetOperationStatus, &request, &response)
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
			Use:   "history",
			Short: "Query the history",
			Long:  ``,
			Args:  cobra.MinimumNArgs(0),
			Run: func(cmd *cobra.Command, args []string) {
				if len(clients) == 0 {
					logrus.Errorf("Invalid client")
					os.Exit(1)
				}

				var nextNumbers []uint32
				for _, arg := range args {
					v, err := strconv.ParseInt(arg, 0, 33)
					if err != nil {
						logrus.Errorf("Could not parse value %q: %v", arg, err)
						os.Exit(1)
					}
					nextNumbers = append(nextNumbers, uint32(v))
				}
				if len(nextNumbers) == 0 {
					nextNumbers = append(nextNumbers, 0)
				}

				for _, client := range clients {
					for _, nextNumber := range nextNumbers {
						logrus.Debugf("Next number: %d", nextNumber)
						request := wire.GetOperationStatusRequest{
							RecordIndex: nextNumber,
						}
						var response wire.GetOperationStatusResponse
						err := client.Raw(wire.FunctionGetOperationStatus, &request, &response)
						if err != nil {
							logrus.Errorf("Error from client: %v", err)
							continue
						}
						if response.Record != nil {
							logrus.Debugf("Record: %+v", *response.Record)
							var person *cobrafile.Person
							var controller, door string
							if controllerList != nil {
								controller, door = controllerList.LookupNameAndDoor(client.ControllerAddress, response.Record.RecordState)
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
								fmt.Printf("%v | Controller: %s | Door: %s | Card ID: %s\n", response.Record.BrushDateTime, controller, door, wire.CardID(response.Record.AreaNumber, response.Record.IDNumber))
							} else {
								fmt.Printf("%v | Controller: %s | Door: %s | Card ID: %s | Name: %s\n", response.Record.BrushDateTime, controller, door, wire.CardID(response.Record.AreaNumber, response.Record.IDNumber), person.Name)
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
						err := client.Raw(wire.FunctionGetBasicInfo, nil, &response)
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
						err := client.Raw(wire.FunctionGetNetworkInfo, request, &response)
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
						err := client.Raw(wire.FunctionGetOperationStatus, &request, &response)
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
									err := client.Raw(wire.FunctionGetOperationStatus, &request, &response)
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
											controller, door = controllerList.LookupNameAndDoor(client.ControllerAddress, response.Record.RecordState)
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
											fmt.Printf("%v | Controller: %s | Door: %s | Card ID: %s\n", response.Record.BrushDateTime, controller, door, wire.CardID(response.Record.AreaNumber, response.Record.IDNumber))
										} else {
											fmt.Printf("%v | Controller: %s | Door: %s | Card ID: %s | Name: %s\n", response.Record.BrushDateTime, controller, door, wire.CardID(response.Record.AreaNumber, response.Record.IDNumber), person.Name)
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
					err := client.Raw(wire.FunctionSetTime, &request, &response)
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

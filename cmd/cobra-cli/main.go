package main

import (
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gitlab.com/doug-manley/cobra-controls/cobrafile"
	"gitlab.com/doug-manley/cobra-controls/wire"
)

func main() {
	var controllerAddress string
	var controllerPort int
	var boardAddressString string
	var boardAddress uint16
	var controllerFile string
	var personnelFile string

	var client *wire.Client
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

			if len(boardAddressString) > 0 {
				v, err := strconv.ParseInt(boardAddressString, 0 /*auto-detect base*/, 17 /*one more than 16 because this is signed*/)
				if err != nil {
					logrus.Errorf("Could not parse board address: %v", err)
					os.Exit(1)
				}
				boardAddress = uint16(v)
				logrus.Infof("Board address: %d (0x%x)", boardAddress, boardAddress)
			}

			if controllerAddress != "" && controllerPort > 0 && boardAddress > 0 {
				client = &wire.Client{
					ControllerAddress: controllerAddress,
					ControllerPort:    controllerPort,
					BoardAddress:      boardAddress,
				}
				logrus.Debugf("Client: %+v", client)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
			os.Exit(1)
		},
	}
	rootCommand.PersistentFlags().StringVar(&controllerAddress, "controller-address", "", "Set the controller address")
	rootCommand.PersistentFlags().IntVar(&controllerPort, "controller-port", 60000, "Set the controller address")
	rootCommand.PersistentFlags().StringVar(&boardAddressString, "board-address", "", "Set the board address (either hexadecimal or decimial)")
	rootCommand.PersistentFlags().StringVar(&controllerFile, "controller-file", "", "Use this CSV file to load the controller information")
	rootCommand.PersistentFlags().StringVar(&personnelFile, "personnel-file", "", "Use this CSV file to load the personnel information")
	rootCommand.PersistentFlags().BoolVar(&verbose, "verbose", false, "Enable verbose output")

	{
		cmd := &cobra.Command{
			Use:   "info",
			Short: "Gather information",
			Long:  ``,
			Run: func(cmd *cobra.Command, args []string) {
				if client == nil {
					logrus.Errorf("Invalid client")
					os.Exit(1)
				}

				{
					var response wire.GetBasicInfoResponse
					err := client.Raw(wire.FunctionGetBasicInfo, nil, &response)
					if err != nil {
						logrus.Errorf("Error: %v", err)
						os.Exit(1)
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
						os.Exit(1)
					}
					logrus.Infof("Response: %+v", response)
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
			Run: func(cmd *cobra.Command, args []string) {
				if client == nil {
					logrus.Errorf("Invalid client")
					os.Exit(1)
				}

				var nextNumber uint32 // If this is zero, then we'll ask for the latest value.
				for batch := 0; ; batch++ {
					if batchCount > 0 {
						if batch >= batchCount {
							break
						}
					}
					if batch > 0 {
						logrus.Infof("Sleeping for %v.", sleepDuration)
						time.Sleep(sleepDuration)
					}

					logrus.Infof("Next number: %d", nextNumber)
					request := wire.GetOperationStatusRequest{
						RecordIndex: nextNumber,
					}
					var response wire.GetOperationStatusResponse
					err := client.Raw(wire.FunctionGetOperationStatus, &request, &response)
					if err != nil {
						logrus.Errorf("Error: %v", err)
						os.Exit(1)
					}
					logrus.Infof("Response: %+v", response)
					if response.RecordCount >= nextNumber {
						if nextNumber > 0 && response.RecordCount >= nextNumber {
							for index := nextNumber; index <= response.RecordCount; index++ {
								logrus.Infof("Geting record %d", index)
								request := wire.GetOperationStatusRequest{
									RecordIndex: index,
								}
								var response wire.GetOperationStatusResponse
								err := client.Raw(wire.FunctionGetOperationStatus, &request, &response)
								if err != nil {
									logrus.Errorf("Error: %v", err)
									os.Exit(1)
								}
								logrus.Infof("Response: %+v", response)
								if response.Record != nil {
									logrus.Infof("Record: %+v", *response.Record)
								}
							}
						}
						nextNumber = response.RecordCount + 1
					}
				}
			},
		}
		cmd.Flags().IntVar(&batchCount, "batch", 10, "How many iterations to run (use 0 for infinite)")
		cmd.Flags().DurationVar(&sleepDuration, "batch-interval", 5*time.Second, "How long to wait between batches")

		rootCommand.AddCommand(cmd)
	}

	err := rootCommand.Execute()
	if err != nil {
		logrus.Errorf("Error: %v", err)
	}
	os.Exit(0)
}

package main

import (
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gitlab.com/doug-manley/cobra-controls/wire"
)

func main() {
	var controllerAddress string
	var controllerPort int
	var boardAddressString string
	var boardAddress uint16
	var client *wire.Client
	verbose := false

	rootCommand := &cobra.Command{
		Use:   "cobra-cli",
		Short: "Command-line tools for Cobra Controls access systems",
		Long:  ``,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if verbose {
				logrus.SetLevel(logrus.DebugLevel)
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
		cmd := &cobra.Command{
			Use:   "monitor",
			Short: "Monitor a door",
			Long:  ``,
			Run: func(cmd *cobra.Command, args []string) {
				if client == nil {
					logrus.Errorf("Invalid client")
					os.Exit(1)
				}

				var lastNumber uint32
				for {
					logrus.Infof("Last number: %d", lastNumber)
					request := wire.GetOperationStatusRequest{
						RecordIndex: lastNumber,
					}
					var response wire.GetOperationStatusResponse
					err := client.Raw(wire.FunctionGetOperationStatus, &request, &response)
					if err != nil {
						logrus.Errorf("Error: %v", err)
						os.Exit(1)
					}
					logrus.Infof("Response: %+v", response)
					if response.RecordCount != lastNumber {
						lastNumber = response.RecordCount
					}

					time.Sleep(1 * time.Second)
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

package main

import (
	"fmt"
	"net"
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
	var conn net.Conn
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
				var err error
				conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", controllerAddress, controllerPort))
				if err != nil {
					logrus.Errorf("Could not connect to controller: %v", err)
					os.Exit(1)
				}
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
				if conn == nil {
					logrus.Errorf("Invalid connection")
					os.Exit(1)
				}

				{
					conn.SetDeadline(time.Now().Add(5 * time.Second))

					envelope := wire.Envelope{
						BoardAddress: boardAddress,
						Function:     0x1082,
					}
					contents, err := wire.Encode(&envelope)
					if err != nil {
						logrus.Errorf("Could not encode envelope: %v", err)
						os.Exit(1)
					}
					bytesWritten, err := conn.Write(contents)
					if err != nil {
						logrus.Errorf("Could not write contents: %v", err)
						os.Exit(1)
					}
					logrus.Infof("Bytes written: %d", bytesWritten)
					if bytesWritten != len(contents) {
						logrus.Errorf("Could not write full contents; wrote %d bytes (expected: %d)", bytesWritten, len(contents))
						os.Exit(1)
					}
				}

				{
					conn.SetDeadline(time.Now().Add(5 * time.Second))

					contents := make([]byte, 1024)
					bytesRead, err := conn.Read(contents)
					if err != nil {
						logrus.Errorf("Could not read contents: %v", err)
						os.Exit(1)
					}
					contents = contents[0:bytesRead]
					logrus.Infof("Bytes read: (%d) %x", bytesRead, contents)

					var envelope wire.Envelope
					err = wire.Decode(contents, &envelope)
					if err != nil {
						logrus.Errorf("Could not read contents: %v", err)
						os.Exit(1)
					}
					logrus.Infof("Response: %x", envelope.Contents)
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
				// TODO
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

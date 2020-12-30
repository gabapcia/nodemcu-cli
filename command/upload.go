package command

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
)

const (
	SOURCE_PATH_FLAG_NAME = "source-path"
	BAUD_RATE_FLAG_NAME   = "baud-rate"
	USB_PORT_FLAG_NAME    = "port"
)

func upload(sourcePath, port string, baudRate int) error {
	sourcePath, err := filepath.Abs(sourcePath)
	if err != nil {
		return fmt.Errorf("Invalid source path")
	}

	files, err := ioutil.ReadDir(sourcePath)
	if err != nil {
		return fmt.Errorf("Error reading the provided source path")
	}

	luaFiles := make([]string, 0)
	for _, file := range files {
		filename := file.Name()
		if !strings.HasSuffix(filename, ".lua") {
			continue
		}

		luaFilePath := filepath.Join(sourcePath, filename)
		luaFiles = append(luaFiles, fmt.Sprintf("%s:%s", luaFilePath, filename))
	}
	if len(luaFiles) == 0 {
		return fmt.Errorf("The source path does not have any .lua file")
	}

	args := []string{"upload"}
	for _, file := range luaFiles {
		args = append(args, file)
	}
	args = append(args, "--restart")

	cmd := exec.Command("nodemcu-uploader", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()

	return err
}

func Upload() *cli.Command {
	return &cli.Command{
		Name:    "upload",
		Usage:   "Upload lua code to NodeMCU",
		Aliases: []string{"up"},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     SOURCE_PATH_FLAG_NAME,
				Aliases:  []string{"s"},
				Required: true,
				Usage:    "Source code base directory",
			},
			&cli.IntFlag{
				Name:    BAUD_RATE_FLAG_NAME,
				Aliases: []string{"b"},
				Value:   115200,
				Usage:   "The speed of the NodeMCU serial connection",
			},
			&cli.StringFlag{
				Name:    USB_PORT_FLAG_NAME,
				Aliases: []string{"p"},
				Value:   "/dev/ttyUSB0",
				Usage:   "NodeMCU USB port",
			},
		},
		Action: func(c *cli.Context) error {
			_, err := exec.LookPath("nodemcu-uploader")
			if err != nil {
				msg := "nodemcu-uploader not found in $PATH. " +
					"Please, use \"pip install nodemcu-uploader\" to install it."
				return fmt.Errorf(msg)
			}

			sourcePath := c.String(SOURCE_PATH_FLAG_NAME)
			baudRate := c.Int(BAUD_RATE_FLAG_NAME)
			port := c.String(USB_PORT_FLAG_NAME)
			return upload(sourcePath, port, baudRate)
		},
	}
}

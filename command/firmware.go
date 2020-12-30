package command

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path"
	"regexp"
	"strings"

	"github.com/urfave/cli/v2"
)

type EnumValue struct {
	Enum     []string
	Default  string
	selected string
}

func (e *EnumValue) Set(value string) error {
	for _, enum := range e.Enum {
		if enum == value {
			e.selected = value
			return nil
		}
	}

	return fmt.Errorf("allowed values are %s", strings.Join(e.Enum, ", "))
}

func (e EnumValue) String() string {
	if e.selected == "" {
		return e.Default
	}
	return e.selected
}

const (
	FLASH_MODE_FLAG_NAME              = "flash-mode"
	FIRMWARE_DIRECTORY_FLAG_NAME      = "directory"
	FORCE_FIRMWARE_DOWNLOAD_FLAG_NAME = "force-download"
)

func createDirectory(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			return fmt.Errorf("Error creating the firmware directory: %s", err)
		}
	}
	return nil
}

func downloadFirmware(path string, forceDownload bool) error {
	if forceDownload {
		if err := os.RemoveAll(path); err != nil {
			return err
		}
		if err := createDirectory(path); err != nil {
			return err
		}
	} else {
		files, err := ioutil.ReadDir(path)
		if err != nil {
			return err
		}

		if len(files) > 0 {
			return nil
		}
	}

	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git not found in $PATH. Please, install git to continue")
	}

	cmd := exec.Command(
		"git",
		"clone",
		"--recurse-submodules",
		"-b",
		"release",
		"https://github.com/nodemcu/nodemcu-firmware.git",
		path,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	return err
}

func selectFirmwareModules(firmwareDir string) error {
	modulesFile := path.Join(firmwareDir, "app", "include", "user_modules.h")

	file, err := os.Open(modulesFile)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(file)

	scanner.Split(bufio.ScanLines)
	var text []string
	modules := []string{"HTTP", "PWM", "SJSON", "TLS"}

	for scanner.Scan() {
		line := scanner.Text()
		for _, module := range modules {
			if strings.HasSuffix(line, module) {
				line = strings.Replace(line, "//", "", 1)
				break
			}
		}
		text = append(text, line)
	}
	file.Close()

	if err := os.Remove(modulesFile); err != nil {
		return err
	}

	file, err = os.Create(modulesFile)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range text {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}

func selectUserConfig(firmwareDir string) error {
	configFile := path.Join(firmwareDir, "app", "include", "user_config.h")
	file, err := os.Open(configFile)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(file)

	scanner.Split(bufio.ScanLines)
	var text []string
	modules := []string{"CLIENT_SSL_ENABLE"}

	for scanner.Scan() {
		line := scanner.Text()
		for _, module := range modules {
			match, _ := regexp.MatchString(`^(//)?#define\s+`, line)
			if match && strings.HasSuffix(line, module) {
				line = strings.Replace(line, "//", "", 1)
				break
			}
		}
		text = append(text, line)
	}
	file.Close()

	if err := os.Remove(configFile); err != nil {
		return err
	}

	file, err = os.Create(configFile)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range text {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}

func buildFirmware(firmwareDir string) error {
	binariesPath := path.Join(firmwareDir, "bin")
	files, err := ioutil.ReadDir(binariesPath)
	if err != nil {
		return err
	}

	binaryFiles := make([]string, 0)
	for _, file := range files {
		filename := file.Name()
		if strings.HasSuffix(filename, ".bin") {
			binaryFiles = append(binaryFiles, filename)
		}
	}

	if len(binaryFiles) > 0 {
		return nil
	}

	cmd := exec.Command("make")
	cmd.Dir = firmwareDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func flashFirmware(firmwareDir, port, flashMode string) error {
	binariesPath := path.Join(firmwareDir, "bin")
	bins := []string{"0x00000", "0x10000"}

	for _, bin := range bins {
		cmd := exec.Command(
			"esptool.py",
			"--port",
			port,
			"write_flash",
			"-fm",
			flashMode,
			bin,
			fmt.Sprintf("%v.bin", bin),
		)
		cmd.Dir = binariesPath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return err
		}
	}

	return nil
}

func firmware(flashMode, port, outputPath string, forceDownload bool) error {
	if err := downloadFirmware(outputPath, forceDownload); err != nil {
		return err
	}

	if err := selectFirmwareModules(outputPath); err != nil {
		return err
	}

	if err := selectUserConfig(outputPath); err != nil {
		return err
	}

	if err := buildFirmware(outputPath); err != nil {
		return err
	}

	if err := flashFirmware(outputPath, port, flashMode); err != nil {
		return err
	}

	return nil
}

func Firmware() *cli.Command {
	currentUser, err := user.Current()
	if err != nil {
		log.Fatalf("Current user could not be detected")
	}
	outputPath := path.Join(currentUser.HomeDir, "nodemcu-firmware")

	return &cli.Command{
		Name:    "firmware",
		Usage:   "Flash firmware to NodeMCU",
		Aliases: []string{"firm"},
		Flags: []cli.Flag{
			&cli.GenericFlag{
				Name:    FLASH_MODE_FLAG_NAME,
				Aliases: []string{"f"},
				Usage:   "NodeMCU flash mode",
				Value: &EnumValue{
					Enum:    []string{"qio", "dio", "dout"},
					Default: "dio",
				},
			},
			&cli.StringFlag{
				Name:    USB_PORT_FLAG_NAME,
				Aliases: []string{"p"},
				Value:   "/dev/ttyUSB0",
				Usage:   "NodeMCU USB port",
			},
			&cli.PathFlag{
				Name:    FIRMWARE_DIRECTORY_FLAG_NAME,
				Aliases: []string{"d"},
				Usage:   "Download firmware directory",
				Value:   outputPath,
			},
			&cli.BoolFlag{
				Name:  FORCE_FIRMWARE_DOWNLOAD_FLAG_NAME,
				Usage: "Clean the existing folder and download new firmware",
			},
		},
		Action: func(c *cli.Context) error {
			var err error
			_, err = exec.LookPath("git")
			if err != nil {
				msg := "\"git\" not found in $PATH."
				return fmt.Errorf(msg)
			}

			_, err = exec.LookPath("make")
			if err != nil {
				msg := "\"make\" not found in $PATH."
				return fmt.Errorf(msg)
			}

			_, err = exec.LookPath("esptool.py")
			if err != nil {
				msg := "\"esptool.py\" not found in $PATH. " +
					"Please, use \"pip install esptool\" to install it."
				return fmt.Errorf(msg)
			}

			outputPath := c.String(FIRMWARE_DIRECTORY_FLAG_NAME)

			if err := createDirectory(outputPath); err != nil {
				return err
			}

			flashMode := c.String(FLASH_MODE_FLAG_NAME)
			port := c.String(USB_PORT_FLAG_NAME)
			forceDownload := c.Bool(FORCE_FIRMWARE_DOWNLOAD_FLAG_NAME)

			return firmware(flashMode, port, outputPath, forceDownload)
		},
	}
}

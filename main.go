package main

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"log"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const applicationName string = "tplink-hs1x-cli"
const applicationVersion string = "v1.0"

type SimpleResponse struct {
	System struct {
		SetRelayState struct {
			ErrCode int `json:"err_code"`
		} `json:"set_relay_state"`
	} `json:"system"`
}

type LedOnOffResponse struct {
	System struct {
		SetLedOff struct {
			ErrCode int `json:"err_code"`
		} `json:"set_led_off"`
	} `json:"system"`
}

type TimeResponse struct {
	Time struct {
		GetTime struct {
			Year    int `json:"year"`
			Month   int `json:"month"`
			Mday    int `json:"mday"`
			Hour    int `json:"hour"`
			Min     int `json:"min"`
			Sec     int `json:"sec"`
			ErrCode int `json:"err_code"`
		} `json:"get_time"`
	} `json:"time"`
}

type WifiScanResponse struct {
	Netif struct {
		GetScaninfo struct {
			ApList []struct {
				Ssid    string `json:"ssid"`
				KeyType int    `json:"key_type"`
			} `json:"ap_list"`
			ErrCode int `json:"err_code"`
		} `json:"get_scaninfo"`
	} `json:"netif"`
}

type SystemInfo struct {
	System struct {
		GetSysinfo struct {
			SwVer      string `json:"sw_ver"`
			HwVer      string `json:"hw_ver"`
			Type       string `json:"type"`
			Model      string `json:"model"`
			Mac        string `json:"mac"`
			DevName    string `json:"dev_name"`
			Alias      string `json:"alias"`
			RelayState int    `json:"relay_state"`
			OnTime     int    `json:"on_time"`
			ActiveMode string `json:"active_mode"`
			Feature    string `json:"feature"`
			Updating   int    `json:"updating"`
			IconHash   string `json:"icon_hash"`
			Rssi       int    `json:"rssi"`
			LedOff     int    `json:"led_off"`
			LongitudeI int    `json:"longitude_i"`
			LatitudeI  int    `json:"latitude_i"`
			HwID       string `json:"hwId"`
			FwID       string `json:"fwId"`
			DeviceID   string `json:"deviceId"`
			OemID      string `json:"oemId"`
			NextAction struct {
				Type    int    `json:"type"`
				ID      string `json:"id"`
				SchdSec int    `json:"schd_sec"`
				Action  int    `json:"action"`
			} `json:"next_action"`
			NtcState int `json:"ntc_state"`
			ErrCode  int `json:"err_code"`
		} `json:"get_sysinfo"`
	} `json:"system"`
}

var (
	// further commands listed here: https://github.com/softScheck/tplink-smartplug/blob/master/tplink-smarthome-commands.txt
	commandList = map[string]string{
		"on":        `{"system":{"set_relay_state":{"state":1}}}`,
		"off":       `{"system":{"set_relay_state":{"state":0}}}`,
		"info":      `{"system":{"get_sysinfo":{}}}`,
		"status":    `{"system":{"get_sysinfo":{}}}`, // same as info, just output is parsed differently
		"wifiscan":  `{"netif":{"get_scaninfo":{"refresh":1}}}`,
		"getaction": `{"schedule":{"get_next_action":null}}`,
		"getrules":  `{"schedule":{"get_rules":null}}`,
		"getaway":   `{"anti_theft":{"get_rules":null}}`,
		"reboot":    `{"system":{"reboot":{"delay":1}}}`,
		"ledoff":    `{"system":{"set_led_off":{"off":1}}}`,
		"ledon":     `{"system":{"set_led_off":{"off":0}}}`,
		"cloudinfo": `{"cnCloud":{"get_info":{}}}`,
		"gettime":   `{"time":{"get_time":{}}}`,
	}

	myDevice string
)

func init() {
	flag.String("config", "config.yaml", "Configuration file: /path/to/file.yaml, default = ./config.yaml")
	flag.String("do", "on", "on, off, info, cloudinfo, ledon, ledoff, wifiscan, getaction, gettime, getrules, getaway, reboot, status (default: \"on\")")
	flag.Bool("debug", false, "Display debugging information")
	flag.Bool("list", false, "Display my devices")
	flag.Bool("displayconfig", false, "Display configuration")
	flag.Bool("help", false, "Display help")
	flag.Bool("version", false, "Display version information")
	flag.Bool("all", false, "For all devices")
	flag.String("device", "", "What device to query, (default: \"all\")")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	err := viper.BindPFlags(pflag.CommandLine)
	checkErr(err)

	if viper.GetBool("help") {
		displayHelp()
		os.Exit(0)
	}

	if viper.GetBool("version") {
		fmt.Println(applicationName + " " + applicationVersion)
		os.Exit(0)
	}

	configdir, configfile := filepath.Split(viper.GetString("config"))

	// set default configuration directory to current directory
	if configdir == "" {
		configdir = "."
	}

	viper.SetConfigType("yaml")
	viper.AddConfigPath(configdir)

	config := strings.TrimSuffix(configfile, ".yaml")
	config = strings.TrimSuffix(config, ".yml")

	viper.SetConfigName(config)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatal("Config file not found")
		} else {
			log.Fatal("Config file was found but another error was discovered: ", err)
		}
	}

	if viper.GetBool("displayconfig") {
		displayConfig()
		os.Exit(0)
	}

	if viper.GetBool("list") {
		displayDevices()
		os.Exit(0)
	}

	if !isCommandValid(viper.GetString("do")) {
		fmt.Printf("\"--do %s\" is an invalid action\n", strings.ToLower(viper.GetString("do")))
		os.Exit(1)
	}

	if viper.GetBool("all") || (len(viper.GetString("device")) == 0) {
		// if "--all" or if default is used, assume "all"
		myDevice = "all"
	} else {

		// check that the device exists
		if _, ok := viper.GetStringMap("devices")[viper.GetString("device")]; ok {
			myDevice = viper.GetString("device")
		} else {
			// device isn't found

			// check if user has manually set "--device all"
			if strings.EqualFold(viper.GetString("device"), "all") {
				myDevice = "all"

			} else {
				// exit out saying device not found
				fmt.Printf("Device %s does not exist, exiting\n", viper.GetString("device"))
				os.Exit(1)
			}
		}

	}

}

func main() {

	// map of ips to device names (aka a reverse of "devices" in config file)
	allDevices := make(map[string]string)
	for deviceName, ip := range viper.GetStringMap("devices") {
		allDevices[ip.(string)] = deviceName
	}

	// list of ips to send commands to
	ips := make([]string, 0)

	// if all devices,
	if strings.EqualFold(myDevice, "all") {
		//fmt.Println("all devices")
		for _, ip := range viper.GetStringMap("devices") {
			ips = append(ips, ip.(string))
		}

	} else {
		ips = append(ips, viper.GetStringMap("devices")[viper.GetString("device")].(string))
	}

	if len(ips) > 1 {
		fmt.Println("Devices to control:")
		for _, ip := range ips {
			fmt.Printf("%s", allDevices[ip])
			if viper.GetBool("debug") {
				fmt.Printf(" %s", ip)
			}
			fmt.Printf("\n")
		}
	}

	for _, ip := range ips {

		if len(ips) > 1 {
			fmt.Printf("\nDevice: %s\n==============\n", allDevices[ip])
		}

		jsonCmd := commandList[strings.ToLower(viper.GetString("do"))]

		data := encrypt(jsonCmd)

		if viper.GetBool("debug") {
			fmt.Printf("Sending: %s\n", jsonCmd)
		}
		reading, err := send(ip, data)
		if err == nil {

			// strip out junk at end of result in preparation for json parsing
			decryptedresponse := decrypt(reading[4:])
			lastinstance := strings.LastIndex(decryptedresponse, "}")
			decryptedresponse = decryptedresponse[:lastinstance] + "}"

			var prettyJSON bytes.Buffer
			error := json.Indent(&prettyJSON, []byte(decryptedresponse), "", " ")
			if error != nil {
				log.Println("JSON parse error: ", error)
			}

			// if standard on and off, show response
			if strings.EqualFold(viper.GetString("do"), "on") || strings.EqualFold(viper.GetString("do"), "off") {
				res := SimpleResponse{}
				json.Unmarshal([]byte(decryptedresponse), &res)

				if len(ips) > 1 {
					fmt.Printf("%s: ", allDevices[ip])
				}

				switch res.System.SetRelayState.ErrCode {
				case 0:
					fmt.Println("OK")
				default:
					fmt.Printf("ERROR:%d\n", res.System.SetRelayState.ErrCode)
				}
			}

			// if LED on and off, show response
			if strings.EqualFold(viper.GetString("do"), "ledon") || strings.EqualFold(viper.GetString("do"), "ledoff") {
				res := LedOnOffResponse{}
				json.Unmarshal([]byte(decryptedresponse), &res)

				if len(ips) > 1 {
					fmt.Printf("%s: ", allDevices[ip])
				}

				switch res.System.SetLedOff.ErrCode {
				case 0:
					fmt.Println("OK")
				default:
					fmt.Printf("ERROR:%d\n", res.System.SetLedOff.ErrCode)
				}
			}

			// show gettime response
			if strings.EqualFold(viper.GetString("do"), "gettime") {
				res := TimeResponse{}
				json.Unmarshal([]byte(decryptedresponse), &res)

				switch res.Time.GetTime.ErrCode {
				case 0:
					fmt.Printf("%04d-%02d-%02d %02d:%02d:%02d\n", res.Time.GetTime.Year, res.Time.GetTime.Month, res.Time.GetTime.Mday, res.Time.GetTime.Hour, res.Time.GetTime.Min, res.Time.GetTime.Sec)
				default:
					fmt.Printf("ERROR:%d\n", res.Time.GetTime.ErrCode)
				}
			}

			// show wifiscan response
			if strings.EqualFold(viper.GetString("do"), "wifiscan") {
				res := WifiScanResponse{}
				json.Unmarshal([]byte(decryptedresponse), &res)

				switch res.Netif.GetScaninfo.ErrCode {
				case 0:
					for _, bbbb := range res.Netif.GetScaninfo.ApList {
						fmt.Printf("%s\n", bbbb.Ssid)
					}
				default:
					fmt.Printf("ERROR:%d\n", res.Netif.GetScaninfo.ErrCode)
				}
			}

			// print the entire json response if info, getaction, getrules, getaway, wificscan
			if viper.GetBool("debug") || strings.EqualFold(viper.GetString("do"), "info") || strings.EqualFold(viper.GetString("do"), "cloudinfo") || strings.EqualFold(viper.GetString("do"), "getaction") || strings.EqualFold(viper.GetString("do"), "getrules") || strings.EqualFold(viper.GetString("do"), "getaway") {

				fmt.Printf("Response: %s\n", string(prettyJSON.Bytes()))
			}

			// print status of a device (on or off)
			if strings.EqualFold(viper.GetString("do"), "status") {
				res := SystemInfo{}
				json.Unmarshal([]byte(decryptedresponse), &res)

				if len(ips) > 1 {
					fmt.Printf("%s: ", allDevices[ip])
				}

				switch res.System.GetSysinfo.RelayState {
				case 0:
					fmt.Println("OFF")
				case 1:
					fmt.Println("ON")
				}
			}

		}
	}
}

// checks errors
func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// decrypts the return message
func decrypt(ciphertext []byte) string {
	n := len(ciphertext)
	key := byte(0xAB)
	var nextKey byte
	for i := 0; i < n; i++ {
		nextKey = ciphertext[i]
		ciphertext[i] = ciphertext[i] ^ key
		key = nextKey
	}
	return string(ciphertext)
}

// encrypts a message to be sent to the device
func encrypt(plaintext string) []byte {
	n := len(plaintext)
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, uint32(n))
	ciphertext := []byte(buf.Bytes())

	key := byte(0xAB)
	payload := make([]byte, n)
	for i := 0; i < n; i++ {
		payload[i] = plaintext[i] ^ key
		key = payload[i]
	}

	for i := 0; i < len(payload); i++ {
		ciphertext = append(ciphertext, payload[i])
	}

	return ciphertext
}

// sends a message to the device
func send(ip string, payload []byte) (data []byte, err error) {
	conn, err := net.Dial("tcp", ip+":9999")
	if err != nil {
		fmt.Println("Cannot connnect to plug:", err)
		data = nil
		return
	}
	defer conn.Close()

	_, err = conn.Write(payload)

	reply := make([]byte, 1024)
	_, err = conn.Read(reply)
	if err != nil {
		fmt.Println("Cannot read data from plug:", err)
	}

	// displays reply payload
	if viper.GetBool("debug") {
		fmt.Printf("Response base64: ")
		fmt.Println(b64.StdEncoding.EncodeToString([]byte(reply)))
	}

	return reply, err
}

// displays help information
func displayHelp() {
	message := `
      --config [file]       Configuration file: /path/to/file.yaml (default: "./config.yaml")
      --debug               Display debug information
      --device [string]     Device to apply "do action" against
      --displayconfig       Display configuration
      --do <action>         on, off, info, cloudinfo, ledon, ledoff, wifiscan, getaction, gettime, getrules, getaway, reboot, status (default: "on")
      --help                Display help
      --list                List devices
      --version             Display version`
	fmt.Println(applicationName + " " + applicationVersion)
	fmt.Println(message)
}

func displayConfig() {
	allmysettings := viper.AllSettings()
	var keys []string
	for k := range allmysettings {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Println("CONFIG:", k, ":", allmysettings[k])
	}
}

func displayDevices() {
	if viper.IsSet("devices") {
		for k, v := range viper.GetStringMap("devices") {
			fmt.Printf("%s     %s\n", k, v)
		}
	} else {
		fmt.Println("no devices found")
	}
}

// checks if a command is valid
func isCommandValid(command string) bool {
	if _, ok := commandList[command]; ok {
		return true
	}

	return false
}

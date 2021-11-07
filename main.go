package main

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/binary"
	"flag"
	"fmt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"log"
	"net"
	"os"
	"sort"
	"strings"
)

const applicationName string = "hs100-cli"
const applicationVersion string = "v0.2"

var (
	// further commands listed here: https://github.com/softScheck/tplink-smartplug/blob/master/tplink-smarthome-commands.txt
	commandList = map[string]string{
		"on":        `{"system":{"set_relay_state":{"state":1}}}`,
		"off":       `{"system":{"set_relay_state":{"state":0}}}`,
		"info":      `{"system":{"get_sysinfo":{}}}`,
		"wifiscan":  `{"netif":{"get_scaninfo":{"refresh":1}}}`,
		"getaction": `{"schedule":{"get_next_action":null}}`,
		"getrules":  `{"schedule":{"get_rules":null}}`,
		"getaway":   `{"anti_theft":{"get_rules":null}}`,
	}
)

func init() {
	flag.String("do", "on", "Some Description")
	flag.Bool("debug", false, "Display debugging information")
	flag.Bool("displayconfig", false, "Display configuration")
	flag.Bool("help", false, "Display help")
	flag.Bool("version", false, "Display version information")
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

	if viper.GetBool("displayconfig") {
		displayConfig()
		os.Exit(0)
	}
}

func main() {
	ip := "192.168.10.127"
	json := commandList[strings.ToLower(viper.GetString("do"))]
	data := encrypt(json)
	reading, err := send(ip, data)
	fmt.Println("send complete")
	if err == nil {
		fmt.Printf("Results=%s\n", decrypt(reading[4:]))
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
		fmt.Println(b64.StdEncoding.EncodeToString([]byte(reply)))
	}

	return reply, err
}

// displays help information
func displayHelp() {
	message := `
      --config [file]       Configuration file: /path/to/file.yaml (default: "./config.yaml")
      --debug               Display debug information
      --displayconfig       Display configuration
      --do <action>         on, off, info, wifiscan, getaction, getrules, getaway (default: "on")
      --help                Display help
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

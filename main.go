package main

import (
	b64 "encoding/base64"
	"flag"
	"fmt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"log"
	"net"
)

var (
	myAction    string
	commandList = map[string]string{
		"off":    "AAAAKtDygfiL/5r31e+UtsWg1Iv5nPCR6LfEsNGlwOLYo4HyhueT9tTu3qPeow==",
		"on":     "AAAAKtDygfiL/5r31e+UtsWg1Iv5nPCR6LfEsNGlwOLYo4HyhueT9tTu36Lfog==",
		"status": "AAAAI9Dw0qHYq9+61/XPtJS20bTAn+yV5o/hh+jK8J7rh+vLtpbr",
	}
)

func init() {
	flag.String("do", "on", "Some Description")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	err := viper.BindPFlags(pflag.CommandLine)

	if err != nil {
		fmt.Println("Error in bindpflags")
	}

	fmt.Printf("myaction=%s\n", commandList[viper.GetString("do")])

}

func main() {

	con, err := net.Dial("tcp", "192.168.10.127:9999")

	checkErr(err)

	defer con.Close()

	msg := commandList[viper.GetString("do")]

	Decoded, _ := b64.StdEncoding.DecodeString(msg)

	//fmt.Printf("Decrypted=%s\n", decryptHS(Decoded))

	_, err = con.Write([]byte(Decoded))

	checkErr(err)

	reply := make([]byte, 1024)

	_, err = con.Read(reply)

	checkErr(err)

	fmt.Println(b64.StdEncoding.EncodeToString([]byte(reply)))
	//fmt.Println(string(reply))

	fmt.Printf("decryptHS=%s\n", decryptHS(reply))
}

func checkErr(err error) {

	if err != nil {

		log.Fatal(err)
	}
}

func decryptHS(ciphertext []byte) string {
	n := len(ciphertext)
	key := byte(0xAB)
	var nextKey byte
	for i := 4; i < n; i++ {
		nextKey = ciphertext[i]
		ciphertext[i] = ciphertext[i] ^ key
		key = nextKey
	}
	return string(ciphertext)
}

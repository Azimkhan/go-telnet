package main

import (
	"log"
	"os"
	"strconv"

	"github.com/Azimkhan/go-telnet/internal/telnet"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("ip address and port must be supplied.")
	}
	_, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}
	timeout := 0
	if len(os.Args) > 3 {
		timeout, err = strconv.Atoi(os.Args[3])
		if err != nil {
			log.Fatal(err)
		}
	}
	telnet.Serve(os.Args[1], os.Args[2], timeout)
}

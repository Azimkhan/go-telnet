package telnet

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Wraps scanner so that text is written into channel
func wrapScanner(scanner *bufio.Scanner, out chan string) {
	for scanner.Scan() {
		// fmt.Println("SCAN", scanner.Err(), scanner.Text())
		out <- scanner.Text()
	}
	err := scanner.Err()
	if err != nil {
		fmt.Println("Scan error", err)
	}
	close(out)
}

// Read routine
func readRoutine(ctx context.Context, conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	scanChan := make(chan string, 1)
	go wrapScanner(scanner, scanChan)
OUTER:
	for {
		select {
		case <-ctx.Done():
			break OUTER
		case msg, ok := <-scanChan:
			if !ok {
				break OUTER
			}
			log.Printf("From server: %s", msg)
		}
	}
	log.Printf("Read routine finished")
}

// Write routine
func writeRoutine(ctx context.Context, conn net.Conn, closeChan <-chan bool) {
	scanner := bufio.NewScanner(os.Stdin)
	scanChan := make(chan string, 1)
	go wrapScanner(scanner, scanChan)
OUTER:
	for {
		select {
		case <-closeChan:
			break OUTER
		case <-ctx.Done():
			break OUTER
		case text, ok := <-scanChan:
			if !ok {
				break OUTER
			}
			log.Printf("To server: %s", text)
			conn.Write([]byte(fmt.Sprintf("%s\n", text)))

		}
	}
	log.Printf("Write routine finished")
}

func Serve(host string, port string, timeout int) {
	dialer := &net.Dialer{}
	ctx := context.Background()
	var cancel context.CancelFunc
	if timeout > 0 {
		log.Printf("Connection timeout is set to %d seconds\n", timeout)
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	} else {
		ctx, cancel = context.WithCancel(ctx)
	}

	conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(host, port))
	if err != nil {
		log.Fatalf("Cannot connect: %v", err)
	}
	defer conn.Close()

	wg := sync.WaitGroup{}
	wg.Add(1)

	// channel that signals write routine when connection closes
	closeChan := make(chan bool, 1)
	go func() {
		defer wg.Done()
		readRoutine(ctx, conn)
		closeChan <- true
		close(closeChan)
	}()

	wg.Add(1)

	go func() {
		defer wg.Done()
		writeRoutine(ctx, conn, closeChan)
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		fmt.Println(sig)
		cancel()
	}()

	wg.Wait()

}

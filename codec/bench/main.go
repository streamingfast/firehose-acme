package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ShinyTrinkets/overseer"
	"github.com/streamingfast/cli"
	"github.com/streamingfast/dmetrics"
	"golang.org/x/exp/maps"
)

func init() {
	overseer.DEFAULT_LINE_BUFFER_SIZE = 100 * 1024 * 1024
}

func main() {
	// The "feeder" process accepts only this variable as input, so it must be set
	payload := os.Getenv("PAYLOAD")

	cli.Ensure(payload != "", "PAYLOAD environment variable must be set")
	cli.Ensure(cli.FileExists(payload), "PAYLOAD environment variable %q does not exist", payload)

	cli.Ensure(len(os.Args) == 3, "Usage: PAYLOAD=<file> go run ./codec/bench <feeder> <experiment>")

	validExperiments := map[string]func(string){
		"viaStdinBinary":                   viaStdinBinary,
		"viaStdinLines":                    viaStdinLines,
		"viaGolang":                        viaGolang,
		"viaGolangAndOverseerOutputStream": viaGolangAndOverseerOutputStream,
		"viaOverseer":                      viaOverseer,
	}

	feeder := os.Args[1]
	cli.Ensure(feeder == "-" || cli.FileExists(feeder), "Feeder binary %q does not exist", feeder)

	experiment, found := validExperiments[os.Args[2]]
	cli.Ensure(found, "No experiments named %q found, valid are %q", os.Args[2], strings.Join(maps.Keys(validExperiments), ", "))

	experiment(feeder)
}

type scannerLineStream struct {
	io.Writer
	scanner *bufio.Scanner
}

func newScannerLineStream() *scannerLineStream {
	reader, writer := io.Pipe()

	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64*1024), 100*1024*1024)

	return &scannerLineStream{
		Writer:  writer,
		scanner: scanner,
	}
}

func viaStdinBinary(string) {
	buffer := make([]byte, 64*1024)

	stdOutBytesCounter := dmetrics.NewLocalRateCounter(time.Second, "bytes")

	go func() {
		for {
			count, err := os.Stdin.Read(buffer)
			cli.NoError(err, "Unable to read stdin")

			stdOutBytesCounter.IncBy(int64(count))
		}
	}()

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for {
			<-ticker.C
			fmt.Printf("%s\n", stdOutBytesCounter)
		}
	}()

	time.Sleep(60 * time.Second)
	fmt.Println("Completed")
}

func viaStdinLines(string) {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 64*1024), 100*1024*1024)

	stdOutBytesCounter := dmetrics.NewLocalRateCounter(time.Second, "bytes")
	go func() {
		for scanner.Scan() {
			stdOutBytesCounter.IncBy(int64(len(scanner.Bytes())))
		}
	}()

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for {
			<-ticker.C

			fmt.Printf("%s\n", stdOutBytesCounter)
		}
	}()

	time.Sleep(60 * time.Second)
	fmt.Println("Completed")
}

func viaGolang(feeder string) {
	cmd := exec.Command(feeder)

	lineStream := newScannerLineStream()
	cmd.Stdout = lineStream

	stdOutBytesCounter := dmetrics.NewLocalRateCounter(time.Second, "bytes")
	go func() {
		for lineStream.scanner.Scan() {
			stdOutBytesCounter.IncBy(int64(len(lineStream.scanner.Bytes())))
		}
	}()

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for {
			<-ticker.C

			fmt.Printf("%s\n", stdOutBytesCounter)
		}
	}()

	cmd.Start()

	go func() {
		time.Sleep(60 * time.Second)
		cmd.Process.Signal(os.Interrupt)
	}()

	cmd.Wait()
	fmt.Println("Completed")
}

func viaGolangAndOverseerOutputStream(feeder string) {
	cmd := exec.Command(feeder)

	lines := make(chan string, 10000)
	lineStream := overseer.NewOutputStream(lines)
	lineStream.SetLineBufferSize(100 * 1024 * 1024)

	cmd.Stdout = lineStream

	stdOutBytesCounter := dmetrics.NewLocalRateCounter(time.Second, "bytes")
	go func() {
		for {
			line, ok := <-lines
			if !ok {
				return
			}

			stdOutBytesCounter.IncBy(int64(len(line)))
		}
	}()

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for {
			<-ticker.C

			fmt.Printf("%s\n", stdOutBytesCounter)
		}
	}()

	cmd.Start()

	go func() {
		time.Sleep(60 * time.Second)
		cmd.Process.Signal(os.Interrupt)
	}()

	cmd.Wait()
	fmt.Println("Completed")
}

func viaOverseer(feeder string) {
	cmd := overseer.NewCmd(feeder, overseer.Options{
		Streaming: true,
	})

	stdOutBytesCounter := dmetrics.NewLocalRateCounter(time.Second, "bytes")
	go func() {
		for {
			line, ok := <-cmd.Stdout
			if !ok {
				return
			}

			stdOutBytesCounter.IncBy(int64(len(line)))
		}
	}()

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for {
			<-ticker.C

			fmt.Printf("%s\n", stdOutBytesCounter)
		}
	}()

	cmd.Start()
	time.Sleep(60 * time.Second)

	cmd.Stop()
	<-cmd.Done()

	fmt.Println("Completed")
}

// func viaOverseerWithCustomStream(feeder string) {
// 	cmd := overseer.NewCmd(feeder, overseer.Options{
// 		Streaming: true,
// 	})

// 	lineStream := newScannerLineStream()
// 	cmd.Stdout = lineStream

// 	stdOutBytesCounter := dmetrics.NewLocalRateCounter(time.Second, "bytes")
// 	go func() {
// 		for {
// 			line, ok := <-cmd.Stdout
// 			if !ok {
// 				return
// 			}

// 			stdOutBytesCounter.IncBy(int64(len(line)))
// 		}
// 	}()

// 	go func() {
// 		ticker := time.NewTicker(1 * time.Second)
// 		for {
// 			<-ticker.C

// 			fmt.Printf("%s\n", stdOutBytesCounter)
// 		}
// 	}()

// 	cmd.Start()
// 	time.Sleep(60 * time.Second)

// 	cmd.Stop()
// 	<-cmd.Done()

// 	fmt.Println("Completed")
// }

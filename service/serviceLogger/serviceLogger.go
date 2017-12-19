package serviceLogger

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/codelingo/lingo/app/util"
	"google.golang.org/grpc/grpclog"

	"github.com/briandowns/spinner"
	"github.com/codelingo/kit/log"
	"github.com/codelingo/lingo/app/util/common/config"
	"github.com/juju/errors"
)

type logger struct {
	log.Logger
	spinning bool
	spin     *spinner.Spinner
}

func New() grpclog.Logger {
	return &logger{
		Logger: log.NewLogfmtLogger(os.Stdout),
		spin:   spinner.New(spinner.CharSets[9], 100*time.Millisecond),
	}
}

// Fatal is equivalent to Print() followed by a call to os.Exit() with a non-zero exit code.
func (l *logger) Fatal(args ...interface{}) {
	l.Print(args...)
	os.Exit(1)
}

// Fatalf is equivalent to Printf() followed by a call to os.Exit() with a non-zero exit code.
func (l *logger) Fatalf(format string, args ...interface{}) {
	l.Printf(format, args...)
	os.Exit(1)
}

// Fatalln is equivalent to Println() followed by a call to os.Exit()) with a non-zero exit code.
func (l *logger) Fatalln(args ...interface{}) {
	l.Println(args...)
	os.Exit(1)
}

// Print prints to the logger.
func (l *logger) Print(args ...interface{}) {
	l.Print(args...)
}

// Overwrite the Printf in grpc logger and implement a spinner
func (l *logger) Printf(format string, args ...interface{}) {
	// grpc connection failed
	if strings.Contains(format, "grpc: addrConn.resetTransport failed to create client transport:") ||
		strings.Contains(format, "transport: http2Client.notifyError got notified that the client transport was broken") {
		if !l.spinning {
			// Start spinner
			l.spinning = true
			l.spin.Start()
			fmt.Println("Connecting...")
			// Try to connect to server
			l.reconnect()
		}
	} else {
		util.Logger.Debugf(format, args...)
	}

}

// Check if server is up
func (l *logger) reconnect() {
	go func() {
		cfg, err := config.Platform()
		if err != nil {
			errors.Trace(err)
		}
		grpcAddr, err := cfg.PlatformAddress()
		if err != nil {
			errors.Trace(err)
		}
		raddr, err := net.ResolveTCPAddr("tcp", grpcAddr)
		if err != nil {
			fmt.Println(err)
		}

		for {
			time.Sleep(time.Second * 2)
			_, err = net.DialTCP("tcp", nil, raddr)

			// Server is up
			if err == nil {
				fmt.Println("Reconnected. Please re-run your command.")
				l.spinning = false
				l.spin.Stop()
				os.Exit(0)
				break
			}
		}
	}()
}

// Println prints to the logger.
func (l *logger) Println(args ...interface{}) {
	l.Println(args...)
}

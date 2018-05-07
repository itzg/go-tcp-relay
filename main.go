package main

import (
	"gopkg.in/alecthomas/kingpin.v2"
	"github.com/sirupsen/logrus"
	"net"
	"strconv"
	"io"
	"bytes"
	"bufio"
	"regexp"
)

var (
	in = kingpin.Flag("in", "Port to bind for incoming connection").Required().Int()

	connectRe = regexp.MustCompile("CONNECT ((.+?)?:.+)")
)

func main() {
	kingpin.Parse()

	listener, err := net.Listen("tcp", net.JoinHostPort("", strconv.Itoa(*in)))
	if err != nil {
		logrus.WithError(err).Fatal("Unable to listen")
	}

	logrus.WithField("port", *in).Info("Listening for connections")
	for {
		conn, err := listener.Accept()
		if err != nil {
			logrus.WithError(err).Fatal("Failed to accept a connection")
		}

		go handleConnection(conn)
	}

}

func handleConnection(connIn net.Conn) {
	defer connIn.Close()

	var inspectionBuffer bytes.Buffer

	teeReader := io.TeeReader(connIn, &inspectionBuffer)
	bufReader := bufio.NewReader(teeReader)
	headerLine, err := bufReader.ReadString('\n')
	if err != nil {
		logrus.WithError(err).Error("Failed to read header line")
		return
	}
	logrus.WithField("headerLine", headerLine).Info("Read header line")

	headerLine = headerLine[:len(headerLine)-1]
	commandMatch := connectRe.FindStringSubmatch(headerLine)
	if commandMatch == nil {
		logrus.WithField("command", headerLine).Error("Invalid command")
		return
	}

	destination := commandMatch[1]
	logrus.WithField("destination", destination).Info("Connecting to endpoint")
	connOut, err := net.Dial("tcp", destination)
	if err != nil {
		logrus.WithError(err).WithField("destination", destination).Error("Unable to connect outward")
		return
	}

	bufferedAmount, err := io.Copy(connOut, &inspectionBuffer)
	if err != nil {
		logrus.WithError(err).Error("Failed to replay inspection buffer")
		return
	}
	logrus.WithField("amount", bufferedAmount).Info("Replayed header inspection buffer")

	copied, err := io.Copy(connOut, connIn)
	if err != nil && err != io.EOF {
		logrus.WithError(err).Error("Copy failed")
	}

	logrus.WithField("amount", copied).Info("Finished relay connection")
}

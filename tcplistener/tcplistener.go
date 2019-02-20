package tcplistener

import (
	"bufio"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/irccloud/go-ircevent"
	"github.com/irccloud/irccat/dispatcher"
	"github.com/juju/loggo"
	"github.com/spf13/viper"
	"net"
	"strings"
)

var log = loggo.GetLogger("TCPListener")

type TCPListener struct {
	socket  net.Listener
	irc     *irc.Connection
	twitter *twitter.Client
}

func New() (*TCPListener, error) {
	var err error

	listener := TCPListener{}
	listener.socket, err = net.Listen("tcp", viper.GetString("tcp.listen"))
	if err != nil {
		return nil, err
	}

	return &listener, nil
}

func (l *TCPListener) Run(irccon *irc.Connection, twittercon *twitter.Client) {
	log.Infof("Listening for TCP requests on %s", viper.GetString("tcp.listen"))
	l.irc = irccon
	l.twitter = twittercon
	go l.acceptConnections()
}

func (l *TCPListener) acceptConnections() {
	for {
		conn, err := l.socket.Accept()
		if err != nil {
			break
		}
		go l.handleConnection(conn)
	}
	l.socket.Close()
}

func (l *TCPListener) handleConnection(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		msg = strings.Trim(msg, "\r\n")
		if len(msg) > 0 {
			dispatcher.Send(l.irc, msg, log, conn.RemoteAddr().String())
			tweet, resp, err := l.twitter.Statuses.Update(msg, nil)
			log.Infof("tweet=%s resp=%s err=%s", tweet, resp, err)
		}
	}
	conn.Close()
}

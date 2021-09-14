// basend on https://eli.thegreenplace.net/2020/graceful-shutdown-of-a-tcp-server-in-go/
package servertcp

import (
	"context"
	"io"
	"log"
	"net"
	"sync"

	"github.com/rancher/remotedialer"
)

type TcpServer struct {
	server   *remotedialer.Server
	target   string
	ClientId string
	listener net.Listener
	quit     chan interface{}
	wg       sync.WaitGroup
}

func NewServer(source string, target string, clientId string, dialer *remotedialer.Server) *TcpServer {
	s := &TcpServer{
		quit:     make(chan interface{}),
		ClientId: clientId,
		server:   dialer,
		target:   target,
	}
	l, err := net.Listen("tcp", source)
	if err != nil {
		log.Println(err)
		return nil
	}
	s.listener = l
	s.wg.Add(1)
	go s.serve()
	return s
}

func (s *TcpServer) Stop() {
	close(s.quit)
	s.listener.Close()
	s.wg.Wait()
}

func (s *TcpServer) serve() {
	defer s.wg.Done()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.quit:
				return
			default:
				log.Println("accept error", err)
			}
		} else {
			s.wg.Add(1)
			go func() {
				s.handleConection(conn)
				s.wg.Done()
			}()
		}
	}
}

func (s *TcpServer) handleConection(c net.Conn) {
	defer c.Close()
	remoteAddr := c.RemoteAddr().String()
	log.Printf("serve tcp destination %s to client source %s via %s", s.target, remoteAddr, s.ClientId)
	cRemote, err := s.server.Dialer(s.ClientId)(context.Background(), "tcp", s.target)
	if err != nil {
		log.Println(err)
		return
	}
	go io.Copy(cRemote, c)
	io.Copy(c, cRemote)
	cRemote.Close()
	log.Printf("%s closed tcp", remoteAddr)
}

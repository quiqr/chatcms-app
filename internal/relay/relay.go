package relay

import (
	"context"
	"errors"
	"io"
	"net"
	"sync"
)

func Pipe(a, b net.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)
	copyHalf := func(to, from net.Conn) {
		defer wg.Done()
		io.Copy(to, from)
		if closer, ok := to.(interface{ CloseWrite() error }); ok {
			closer.CloseWrite()
		} else {
			to.Close()
		}
	}
	go copyHalf(a, b)
	go copyHalf(b, a)
	wg.Wait()
	a.Close()
	b.Close()
}

func Serve(ctx context.Context, listener net.Listener, network, address string) error {
	go func() {
		<-ctx.Done()
		listener.Close()
	}()
	for {
		incoming, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, net.ErrClosed) {
				return nil
			}
			return err
		}
		go func() {
			var dialer net.Dialer
			outgoing, err := dialer.DialContext(ctx, network, address)
			if err != nil {
				incoming.Close()
				return
			}
			Pipe(incoming, outgoing)
		}()
	}
}

func Forward(ctx context.Context, listenNetwork, listenAddress, network, address string) (net.Listener, error) {
	listener, err := net.Listen(listenNetwork, listenAddress)
	if err != nil {
		return nil, err
	}
	go Serve(ctx, listener, network, address)
	return listener, nil
}

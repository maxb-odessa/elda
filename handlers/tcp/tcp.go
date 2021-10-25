package tcp

import (
	"net"

	"elda/def"
	l "elda/log"
)

// any name
type handler struct {
	// mandatory fields
	name string
	typ  int

	// optional, internal only
	dataCh chan []byte
	doneCh chan bool
}

// register us
func Register() *handler {
	return &handler{
		name: "tcp",
		typ:  def.HANDLER_TYPE_ACTION,
	}
}

func (self *handler) Init(vars map[string]string) error {

	var remote *net.TCPAddr

	if addr, err := def.GetStrVar(vars, "remote"); err != nil {
		return err
	} else if remote, err = net.ResolveTCPAddr("tcp", addr); err != nil {
		return err
	}

	self.dataCh = make(chan []byte, def.ACT_CHAN_LEN)
	self.doneCh = make(chan bool)

	// start connector and sender
	go func() {

		var conn *net.TCPConn
		var err error

		for {

			select {
			case _ = <-self.doneCh:
				return
			case data := <-self.dataCh:
				if conn == nil {
					if conn, err = net.DialTCP("tcp", nil, remote); err != nil {
						l.Err("Connection to %s failed: %s", err)
						continue
					}
					conn.SetKeepAlive(true)
				}
				if _, err = conn.Write(data); err != nil {
					l.Err("Write to %s failed: %s", err)
					conn.Close()
					conn = nil
				}
			} // select...

		} // for...

	}()

	return nil
}

func (self *handler) Name() string {
	return self.name
}

func (self *handler) Type() int {
	return self.typ
}

// not applicable
func (self *handler) Pull() (string, error) {
	return "", nil
}

func (self *handler) Push(s string) error {
	select {
	case self.dataCh <- []byte(s):
	default:
		l.Warn("remote tcp data chan is full")
	}
	return nil
}

func (self *handler) Done() {
	self.doneCh <- true
}

package mitm

import (
	"bufio"
	"fmt"
	yaklog "github.com/yaklang/yaklang/common/log"
	"net"
	"socks2https/pkg/color"
)

func NewServerHelloDone(version uint16) *Record {
	handshake := &Handshake{
		HandshakeType: HandshakeTypeServerHelloDone,
		Length:        0,
	}
	handshakeRaw := handshake.GetRaw()
	//yaklog.Debugf("handshake raw: %s", color.SetColor(color.RED_COLOR_TYPE, fmt.Sprintf("%v", handshakeRaw)))
	return &Record{
		ContentType: ContentTypeHandshake,
		Version:     version,
		Length:      uint16(len(handshakeRaw)),
		Handshake:   *handshake,
		Fragment:    handshakeRaw,
	}
}

var WriteServerHelloDone = HandleRecord(func(reader *bufio.Reader, conn net.Conn, ctx *Context) error {
	tamplate := fmt.Sprintf("%s [%s] [%s]", ctx.Mitm2ClientLog, color.SetColor(color.YELLOW_COLOR_TYPE, "Handshake"), color.SetColor(color.RED_COLOR_TYPE, "Server Hello Done"))
	record := NewServerHelloDone(ctx.Version)
	ctx.HandshakeMessages = append(ctx.HandshakeMessages, record.Fragment)
	if _, err := conn.Write(record.GetRaw()); err != nil {
		return fmt.Errorf("%s Write ServerHelloDone Failed : %v", tamplate, err)
	}
	yaklog.Infof("%s Write ServerHelloDone Successfully.", tamplate)
	return nil
})

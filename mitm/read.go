package mitm

import (
	"bufio"
	"encoding/binary"
	"fmt"
	yaklog "github.com/yaklang/yaklang/common/log"
	"net"
	"net/http"
	"socks2https/pkg/comm"
	"strings"
)

func HeadProtocol(conn net.Conn, ctx *Context) error {
	reader := bufio.NewReader(conn)
	header, err := reader.Peek(7)
	if err != nil {
		return fmt.Errorf("read Protocol Header failed : %v", err)
	}
	maybeTLS := false
	if _, ok := ContentType[header[0]]; ok {
		maybeTLS = true
	} else if httpMethod, ok := HttpMethod[strings.TrimSpace(string(header))]; ok {
		if httpMethod == http.MethodConnect {
			yaklog.Infof("%s %s", ctx.LogTamplate, comm.SetColor(comm.RED_BG_COLOR_TYPE, comm.SetColor(comm.YELLOW_COLOR_TYPE, "use HTTP CONNECT Connection")))
		}
		yaklog.Infof("%s %s", ctx.LogTamplate, comm.SetColor(comm.RED_COLOR_TYPE, "use HTTP Connection"))
		return HTTPMITM(reader, conn)
	}
	if maybeTLS {
		version := binary.BigEndian.Uint16(header[1:3])
		if version >= VersionSSL30 && version <= VersionTLS13 {
			yaklog.Infof("%s %s", ctx.LogTamplate, comm.SetColor(comm.YELLOW_BG_COLOR_TYPE, comm.SetColor(comm.RED_COLOR_TYPE, "use TSL Connection")))
			TLSMITM(reader, conn, ctx)
			return nil
		}
	}
	yaklog.Infof("%s Client use TCP connection", ctx.LogTamplate)
	return TCPMITM(reader, conn, ctx)
}

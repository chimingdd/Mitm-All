package socks

import (
	"bufio"
	"fmt"
	yaklog "github.com/yaklang/yaklang/common/log"
	"net"
	"socks2https/pkg/comm"
	"socks2https/setting"
)

const (
	HTTP_PROTOCOL  int = 0
	HTTPS_PROTOCOL int = 1
	TCP_PROTOCOL   int = 2

	CONNECT_CMD       byte = 0x01
	BIND_CMD          byte = 0x02
	UDP_ASSOCIATE_CMD byte = 0x03

	RESERVED byte = 0x00

	IPV4_ATYPE byte = 0x01
	FQDN_ATYPE byte = 0x03
	IPV6_ATYPE byte = 0x04

	SUCCEEDED_REP                    byte = 0x00
	GENERAL_SOCKS_SERVER_FAILURE_REP byte = 0x01
	CONNECTION_NOT_ALLOWED_REP       byte = 0x02
	NETWORK_UNREACHABLE_REP          byte = 0x03
	HOST_UNREACHABLE_REP             byte = 0x04
	CONNECTION_REFUSED_REP           byte = 0x05
	TTL_EXPIRED_REP                  byte = 0x06
	COMMAND_NOT_SUPPORTED_REP        byte = 0x07
	ADDRESS_TYPE_NOT_SUPPORTED_REP   byte = 0x08
)

func runcmd(tag string, readWriter *bufio.ReadWriter) (byte, string, error) {
	// 客户端请求包
	// +-----+-----+-------+------+----------+----------+
	// | VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
	// +-----+-----+-------+------+----------+----------+
	// |  1  |  1  | X'00' |  1   | Variable |    2     |
	// +-----+-----+-------+------+----------+----------+
	buf := make([]byte, 4)
	if _, err := readWriter.Read(buf); err != nil {
		return CONNECT_CMD, "", fmt.Errorf("%s read VER CMD RSV ATYP failed : %v", tag, err)
	}
	ver, cmd, rsv, aTyp := buf[0], buf[1], buf[2], buf[3]
	yaklog.Debugf("%s VER : %v , CMD : %v , RSA : %v , ATYP : %v", tag, ver, cmd, rsv, aTyp)
	if ver != SOCKS5_VERSION {
		if _, err := readWriter.Write([]byte{SOCKS5_VERSION, GENERAL_SOCKS_SERVER_FAILURE_REP, RESERVED, IPV4_ATYPE, 0, 0, 0, 0, 0, 0}); err != nil {
			return CONNECT_CMD, "", fmt.Errorf("%s write cmd to Client failed : %v", tag, err)
		} else if err = readWriter.Flush(); err != nil {
			return CONNECT_CMD, "", fmt.Errorf("%s flush cmd failed : %v", tag, err)
		}
		return CONNECT_CMD, "", fmt.Errorf("%s not support socks version : %v", tag, ver)
	} else if cmd != CONNECT_CMD {
		if _, err := readWriter.Write([]byte{SOCKS5_VERSION, COMMAND_NOT_SUPPORTED_REP, RESERVED, IPV4_ATYPE, 0, 0, 0, 0, 0, 0}); err != nil {
			return CONNECT_CMD, "", fmt.Errorf("%s write cmd to Client failed : %v", tag, err)
		} else if err = readWriter.Flush(); err != nil {
			return CONNECT_CMD, "", fmt.Errorf("%s flush cmd failed : %v", tag, err)
		}
		return CONNECT_CMD, "", fmt.Errorf("%s not support command : %v", tag, cmd)
	} else if rsv != RESERVED {
		if _, err := readWriter.Write([]byte{SOCKS5_VERSION, CONNECTION_NOT_ALLOWED_REP, RESERVED, IPV4_ATYPE, 0, 0, 0, 0, 0, 0}); err != nil {
			return CONNECT_CMD, "", fmt.Errorf("%s write cmd to Client failed : %v", tag, err)
		} else if err = readWriter.Flush(); err != nil {
			return CONNECT_CMD, "", fmt.Errorf("%s flush cmd failed : %v", tag, err)
		}
		return CONNECT_CMD, "", fmt.Errorf("%s invail reserved : %v", tag, rsv)
	}
	var host string
	var aLen byte = 0x00
	switch aTyp {
	case IPV6_ATYPE:
		buf = make([]byte, net.IPv6len)
		fallthrough
	case IPV4_ATYPE:
		if _, err := readWriter.Read(buf); err != nil {
			if _, err = readWriter.Write([]byte{SOCKS5_VERSION, GENERAL_SOCKS_SERVER_FAILURE_REP, RESERVED, IPV4_ATYPE, 0, 0, 0, 0, 0, 0}); err != nil {
				return CONNECT_CMD, "", fmt.Errorf("%s write cmd to Client failed : %v", tag, err)
			} else if err = readWriter.Flush(); err != nil {
				return CONNECT_CMD, "", fmt.Errorf("%s flush cmd failed : %v", tag, err)
			}
			return CONNECT_CMD, "", fmt.Errorf("%s read Target IP failed : %v", tag, err)
		}
		host = net.IP(buf).String()
	case FQDN_ATYPE:
		if _, err := readWriter.Read(buf[:1]); err != nil {
			if _, err = readWriter.Write([]byte{SOCKS5_VERSION, GENERAL_SOCKS_SERVER_FAILURE_REP, RESERVED, IPV4_ATYPE, 0, 0, 0, 0, 0, 0}); err != nil {
				return CONNECT_CMD, "", fmt.Errorf("%s write cmd to Client failed : %v", tag, err)
			} else if err = readWriter.Flush(); err != nil {
				return CONNECT_CMD, "", fmt.Errorf("%s flush cmd failed : %v", tag, err)
			}
			return CONNECT_CMD, "", fmt.Errorf("%s read ALEN failed : %v", tag, err)
		}
		aLen = buf[0]
		yaklog.Debugf("%s ALEN : %v", tag, aLen)
		if aLen > net.IPv4len {
			buf = make([]byte, aLen)
		}
		if _, err := readWriter.Read(buf[:aLen]); err != nil {
			if _, err = readWriter.Write([]byte{SOCKS5_VERSION, GENERAL_SOCKS_SERVER_FAILURE_REP, RESERVED, IPV4_ATYPE, 0, 0, 0, 0, 0, 0}); err != nil {
				return CONNECT_CMD, "", fmt.Errorf("%s write cmd to Client failed : %v", tag, err)
			} else if err = readWriter.Flush(); err != nil {
				return CONNECT_CMD, "", fmt.Errorf("%s flush cmd failed : %v", tag, err)
			}
			return CONNECT_CMD, "", fmt.Errorf("%s read Target FQDN failed : %v", tag, err)
		}
		host = string(buf[:aLen])
	default:
		if _, err := readWriter.Write([]byte{SOCKS5_VERSION, ADDRESS_TYPE_NOT_SUPPORTED_REP, RESERVED, IPV4_ATYPE, 0, 0, 0, 0, 0, 0}); err != nil {
			return CONNECT_CMD, "", fmt.Errorf("%s write cmd to Client failed : %v", tag, err)
		} else if err = readWriter.Flush(); err != nil {
			return CONNECT_CMD, "", fmt.Errorf("%s flush cmd failed : %v", tag, err)
		}
		return CONNECT_CMD, "", fmt.Errorf("%s not support address type : %v", tag, aTyp)
	}
	if _, err := readWriter.Read(buf[:2]); err != nil {
		if _, err = readWriter.Write([]byte{SOCKS5_VERSION, ADDRESS_TYPE_NOT_SUPPORTED_REP, RESERVED, IPV4_ATYPE, 0, 0, 0, 0, 0, 0}); err != nil {
			return CONNECT_CMD, "", fmt.Errorf("%s write cmd to Client failed : %v", tag, err)
		} else if err = readWriter.Flush(); err != nil {
			return CONNECT_CMD, "", fmt.Errorf("%s flush cmd failed : %v", tag, err)
		}
		return CONNECT_CMD, host, fmt.Errorf("%s read Target Port failed : %v", tag, err)
	}
	port := (uint16(buf[0]) << 8) + uint16(buf[1])
	addr := fmt.Sprintf("%s:%d", host, port)
	yaklog.Infof("%s Target address : %s", tag, comm.SetColor(comm.GREEN_COLOR_TYPE, addr))
	// 服务端响应包
	// +-----+-----+-------+------+----------+----------+
	// | VER | REP |  RSV  | ATYP | BND.ADDR | BND.PORT |
	// +-----+-----+-------+------+----------+----------+
	// |  1  |  1  |   1   |   1  | Variable |    2     |
	// +-----+-----+-------+------+----------+----------+
	resp := []byte{SOCKS5_VERSION, SUCCEEDED_REP, RESERVED}
	if setting.Bound {
		resp = append(resp, aTyp)
		if aLen != 0x00 {
			resp = append(resp, aLen)
		}
		if _, err := readWriter.Write(append(append(resp, host...), buf[:2]...)); err != nil {
			return CONNECT_CMD, addr, fmt.Errorf("%s write cmd to Client failed : %v", tag, err)
		} else if err = readWriter.Flush(); err != nil {
			return CONNECT_CMD, "", fmt.Errorf("%s flush cmd failed : %v", tag, err)
		}
		return CONNECT_CMD, addr, nil
	}
	if _, err := readWriter.Write(append(resp, IPV4_ATYPE, 0, 0, 0, 0, 0, 0)); err != nil {
		return CONNECT_CMD, addr, fmt.Errorf("%s write cmd to Client failed : %v", tag, err)
	} else if err = readWriter.Flush(); err != nil {
		return CONNECT_CMD, "", fmt.Errorf("%s flush cmd failed : %v", tag, err)
	}
	return CONNECT_CMD, addr, nil
}

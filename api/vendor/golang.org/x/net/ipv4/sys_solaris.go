// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ipv4

import (
	"net"
	"syscall"
	"unsafe"

	"golang.org/x/net/internal/iana"
	"golang.org/x/net/internal/socket"

	"golang.org/x/sys/unix"
)

const sockoptReceiveInterface = unix.IP_RECVIF

var (
	ctlOpts = [ctlMax]ctlOpt{
		ctlTTL:        {unix.IP_RECVTTL, 4, marshalTTL, parseTTL},
		ctlPacketInfo: {unix.IP_PKTINFO, sizeofInetPktinfo, marshalPacketInfo, parsePacketInfo},
	}

	sockOpts = map[int]sockOpt{
		ssoTOS:                {Option: socket.Option{Level: iana.ProtocolIP, Name: unix.IP_TOS, Len: 4}},
		ssoTTL:                {Option: socket.Option{Level: iana.ProtocolIP, Name: unix.IP_TTL, Len: 4}},
		ssoMulticastTTL:       {Option: socket.Option{Level: iana.ProtocolIP, Name: unix.IP_MULTICAST_TTL, Len: 1}},
		ssoMulticastInterface: {Option: socket.Option{Level: iana.ProtocolIP, Name: unix.IP_MULTICAST_IF, Len: 4}},
		ssoMulticastLoopback:  {Option: socket.Option{Level: iana.ProtocolIP, Name: unix.IP_MULTICAST_LOOP, Len: 1}},
		ssoReceiveTTL:         {Option: socket.Option{Level: iana.ProtocolIP, Name: unix.IP_RECVTTL, Len: 4}},
		ssoPacketInfo:         {Option: socket.Option{Level: iana.ProtocolIP, Name: unix.IP_RECVPKTINFO, Len: 4}},
		ssoHeaderPrepend:      {Option: socket.Option{Level: iana.ProtocolIP, Name: unix.IP_HDRINCL, Len: 4}},
		ssoJoinGroup:          {Option: socket.Option{Level: iana.ProtocolIP, Name: unix.MCAST_JOIN_GROUP, Len: sizeofGroupReq}, typ: ssoTypeGroupReq},
		ssoLeaveGroup:         {Option: socket.Option{Level: iana.ProtocolIP, Name: unix.MCAST_LEAVE_GROUP, Len: sizeofGroupReq}, typ: ssoTypeGroupReq},
		ssoJoinSourceGroup:    {Option: socket.Option{Level: iana.ProtocolIP, Name: unix.MCAST_JOIN_SOURCE_GROUP, Len: sizeofGroupSourceReq}, typ: ssoTypeGroupSourceReq},
		ssoLeaveSourceGroup:   {Option: socket.Option{Level: iana.ProtocolIP, Name: unix.MCAST_LEAVE_SOURCE_GROUP, Len: sizeofGroupSourceReq}, typ: ssoTypeGroupSourceReq},
		ssoBlockSourceGroup:   {Option: socket.Option{Level: iana.ProtocolIP, Name: unix.MCAST_BLOCK_SOURCE, Len: sizeofGroupSourceReq}, typ: ssoTypeGroupSourceReq},
		ssoUnblockSourceGroup: {Option: socket.Option{Level: iana.ProtocolIP, Name: unix.MCAST_UNBLOCK_SOURCE, Len: sizeofGroupSourceReq}, typ: ssoTypeGroupSourceReq},
	}
)

func (pi *inetPktinfo) setIfindex(i int) {
	pi.Ifindex = uint32(i)
}

func (gr *groupReq) setGroup(grp net.IP) {
	sa := (*sockaddrInet)(unsafe.Pointer(uintptr(unsafe.Pointer(gr)) + 4))
	sa.Family = syscall.AF_INET
	copy(sa.Addr[:], grp)
}

func (gsr *groupSourceReq) setSourceGroup(grp, src net.IP) {
	sa := (*sockaddrInet)(unsafe.Pointer(uintptr(unsafe.Pointer(gsr)) + 4))
	sa.Family = syscall.AF_INET
	copy(sa.Addr[:], grp)
	sa = (*sockaddrInet)(unsafe.Pointer(uintptr(unsafe.Pointer(gsr)) + 260))
	sa.Family = syscall.AF_INET
	copy(sa.Addr[:], src)
}

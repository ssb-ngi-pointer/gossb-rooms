// SPDX-License-Identifier: MIT

package alias

import (
	"net"

	"github.com/ssb-ngi-pointer/go-ssb-room/roomdb"

	kitlog "github.com/go-kit/kit/log"
	"go.cryptoscope.co/muxrpc/v2"
	"go.cryptoscope.co/muxrpc/v2/typemux"

	"github.com/ssb-ngi-pointer/go-ssb-room/internal/maybemuxrpc"
	refs "go.mindeco.de/ssb-refs"
)

const name = "alias"

var method muxrpc.Method = muxrpc.Method{name}

type plugin struct {
	h   muxrpc.Handler
	log kitlog.Logger
}

func (plugin) Name() string              { return name }
func (plugin) Method() muxrpc.Method     { return method }
func (p plugin) Handler() muxrpc.Handler { return p.h }
func (plugin) Authorize(net.Conn) bool   { return true }

func New(log kitlog.Logger, self refs.FeedRef, aliasDB roomdb.AliasService) maybemuxrpc.Plugin {
	mux := typemux.New(log)

	var h = new(handler)
	h.self = self
	h.logger = log
	h.db = aliasDB
	mux.RegisterAsync(append(method, "register"), typemux.AsyncFunc(h.register))

	return plugin{
		h: &mux,
	}
}

// SPDX-License-Identifier: MIT

package server

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ssb-ngi-pointer/go-ssb-room/internal/network"
	"github.com/ssb-ngi-pointer/go-ssb-room/roomstate"
	refs "go.mindeco.de/ssb-refs"

	kitlog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"go.cryptoscope.co/muxrpc/v2"
)

type handler struct {
	logger kitlog.Logger
	self   refs.FeedRef

	state *roomstate.Manager
}

func (h *handler) isRoom(context.Context, *muxrpc.Request) (interface{}, error) {
	level.Debug(h.logger).Log("called", "isRoom")
	return true, nil
}

func (h *handler) ping(context.Context, *muxrpc.Request) (interface{}, error) {
	now := time.Now().UnixNano() / 1000
	level.Debug(h.logger).Log("called", "ping")
	return now, nil
}

func (h *handler) announce(_ context.Context, req *muxrpc.Request) (interface{}, error) {
	level.Debug(h.logger).Log("called", "announce")
	ref, err := network.GetFeedRefFromAddr(req.RemoteAddr())
	if err != nil {
		return nil, err
	}

	h.state.AddEndpoint(*ref, req.Endpoint())

	return false, nil
}

func (h *handler) leave(_ context.Context, req *muxrpc.Request) (interface{}, error) {
	ref, err := network.GetFeedRefFromAddr(req.RemoteAddr())
	if err != nil {
		return nil, err
	}

	h.state.Remove(*ref)

	return false, nil
}

func (h *handler) endpoints(_ context.Context, req *muxrpc.Request, snk *muxrpc.ByteSink) error {
	level.Debug(h.logger).Log("called", "endpoints")

	toPeer := newForwarder(snk)

	// for future updates
	h.state.Register(toPeer)

	ref, err := network.GetFeedRefFromAddr(req.RemoteAddr())
	if err != nil {
		return err
	}

	has := h.state.AlreadyAdded(*ref, req.Endpoint())
	if !has {
		// just send the current state to the new peer
		toPeer.Update(h.state.List())
	}
	return nil
}

type updateForwarder struct {
	snk *muxrpc.ByteSink
	enc *json.Encoder
}

func newForwarder(snk *muxrpc.ByteSink) updateForwarder {
	enc := json.NewEncoder(snk)
	snk.SetEncoding(muxrpc.TypeJSON)
	return updateForwarder{
		snk: snk,
		enc: enc,
	}
}

func (uf updateForwarder) Update(members []string) error {
	return uf.enc.Encode(members)
}

func (uf updateForwarder) Close() error {
	return uf.snk.Close()
}

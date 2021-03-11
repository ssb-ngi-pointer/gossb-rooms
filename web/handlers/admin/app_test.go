// SPDX-License-Identifier: MIT

package admin

import (
	"context"
	"math/rand"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"go.mindeco.de/http/render"
	"go.mindeco.de/http/tester"
	"go.mindeco.de/logging/logtest"

	"github.com/ssb-ngi-pointer/go-ssb-room/roomdb"
	"github.com/ssb-ngi-pointer/go-ssb-room/roomdb/mockdb"
	"github.com/ssb-ngi-pointer/go-ssb-room/roomstate"
	"github.com/ssb-ngi-pointer/go-ssb-room/web"
	"github.com/ssb-ngi-pointer/go-ssb-room/web/router"
	"github.com/ssb-ngi-pointer/go-ssb-room/web/user"
)

type testSession struct {
	Mux    *http.ServeMux
	Client *tester.Tester
	Router *mux.Router

	AllowListDB *mockdb.FakeAllowListService
	PinnedDB    *mockdb.FakePinnedNoticesService
	NoticeDB    *mockdb.FakeNoticesService
	InvitesDB   *mockdb.FakeInviteService

	User *roomdb.User

	Domain string

	RoomState *roomstate.Manager
}

func newSession(t *testing.T) *testSession {
	var ts testSession

	// fake dbs
	ts.AllowListDB = new(mockdb.FakeAllowListService)
	ts.PinnedDB = new(mockdb.FakePinnedNoticesService)
	ts.NoticeDB = new(mockdb.FakeNoticesService)
	ts.InvitesDB = new(mockdb.FakeInviteService)

	log, _ := logtest.KitLogger("admin", t)
	ctx := context.TODO()
	ts.RoomState = roomstate.NewManager(ctx, log)

	ts.Router = router.CompleteApp()

	ts.Domain = randomString(10)

	// fake user
	ts.User = &roomdb.User{
		ID:   1234,
		Name: "room mate",
	}

	// setup rendering

	// TODO: make testing utils and move these there
	testFuncs := web.TemplateFuncs(ts.Router)
	testFuncs["i18n"] = func(msgID string) string { return msgID }
	testFuncs["i18npl"] = func(msgID string, count int, _ ...interface{}) string {
		if count == 1 {
			return msgID + "Singular"
		}
		return msgID + "Plural"
	}
	testFuncs["current_page_is"] = func(routeName string) bool {
		return true
	}

	testFuncs["is_logged_in"] = func() *roomdb.User { return ts.User }
	testFuncs["urlToNotice"] = func(name string) string { return "" }

	r, err := render.New(web.Templates,
		render.SetLogger(log),
		render.BaseTemplates("base.tmpl", "menu.tmpl"),
		render.AddTemplates(append(HTMLTemplates, "error.tmpl")...),
		render.ErrorTemplate("error.tmpl"),
		render.FuncMap(testFuncs),
	)
	if err != nil {
		t.Fatal(errors.Wrap(err, "setup: render init failed"))
	}

	ts.Mux = http.NewServeMux()

	handler := Handler(
		ts.Domain,
		r,
		ts.RoomState,
		Databases{
			AllowList:     ts.AllowListDB,
			Invites:       ts.InvitesDB,
			Notices:       ts.NoticeDB,
			PinnedNotices: ts.PinnedDB,
		},
	)

	handler = user.MiddlewareForTests(ts.User)(handler)

	ts.Mux.Handle("/", handler)

	ts.Client = tester.New(ts.Mux, t)

	return &ts
}

// utils

func randomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

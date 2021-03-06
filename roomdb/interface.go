// SPDX-License-Identifier: MIT

// Package roomdb implements all the persisted database needs of the room server.
// This includes authentication, allow/deny list managment, invite and alias creation and also the notice content for the CMS.
//
// The interfaces defined here are implemented twice. Once in SQLite for production and once as mocks for testing, generated by counterfeiter (https://github.com/maxbrunsfeld/counterfeiter).
//
// See the package documentation of roomdb/sqlite for how to update it.
// It's important not to use the types generated by sqlboiler (sqlite/models) in the argument and return values of the interfaces here.
// This would leak details of the internal implementation of the roomdb/sqlite package and we want to have full control over how these interfaces can be used.
package roomdb

import (
	"context"

	"go.mindeco.de/http/auth"
	refs "go.mindeco.de/ssb-refs"
)

// AuthFallbackService might be helpful for scenarios where one lost access to his ssb device or key
type AuthFallbackService interface {
	auth.Auther

	Create(ctx context.Context, user string, password []byte) (int64, error)
	GetByID(ctx context.Context, uid int64) (*User, error)
}

// AuthWithSSBService defines functions needed for the challange/response system of sign-in with ssb
type AuthWithSSBService interface{}

// AllowListService changes the lists of people that are allowed to get into the room
type AllowListService interface {
	// Add adds the feed to the list.
	Add(context.Context, refs.FeedRef) error

	// HasFeed returns true if a feed is on the list.
	HasFeed(context.Context, refs.FeedRef) bool

	// HasFeed returns true if a feed is on the list.
	HasID(context.Context, int64) bool

	// GetByID returns the list entry for that ID or an error
	GetByID(context.Context, int64) (ListEntry, error)

	// List returns a list of all the feeds.
	List(context.Context) (ListEntries, error)

	// RemoveFeed removes the feed from the list.
	RemoveFeed(context.Context, refs.FeedRef) error

	// RemoveID removes the feed for the ID from the list.
	RemoveID(context.Context, int64) error
}

// AliasService manages alias handle registration and lookup
type AliasService interface{}

// InviteService manages creation and consumption of invite tokens for joining the room.
type InviteService interface {
	// Create creates a new invite for a new member. It returns the token or an error.
	// createdBy is user ID of the admin or moderator who created it.
	// aliasSuggestion is optional (empty string is fine) but can be used to disambiguate open invites. (See https://github.com/ssb-ngi-pointer/rooms2/issues/21)
	Create(ctx context.Context, createdBy int64, aliasSuggestion string) (string, error)

	// Consume checks if the passed token is still valid.
	// If it is it adds newMember to the members of the room and invalidates the token.
	// If the token isn't valid, it returns an error.
	Consume(ctx context.Context, token string, newMember refs.FeedRef) (Invite, error)

	// GetByToken returns the Invite if one for that token exists, or an error
	GetByToken(ctx context.Context, token string) (Invite, error)

	// GetByToken returns the Invite if one for that ID exists, or an error
	GetByID(ctx context.Context, id int64) (Invite, error)

	// List returns a list of all the valid invites
	List(ctx context.Context) ([]Invite, error)

	// Revoke removes a active invite and invalidates it for future use.
	Revoke(ctx context.Context, id int64) error
}

// PinnedNoticesService allows an admin to assign Notices to specific placeholder pages.
// like updates, privacy policy, code of conduct
type PinnedNoticesService interface {
	// List returns a list of all the pinned notices with their corrosponding notices and languges
	List(context.Context) (PinnedNotices, error)

	// Set assigns a fixed page name to an page ID and a language to allow for multiple translated versions of the same page.
	Set(ctx context.Context, name PinnedNoticeName, id int64) error

	// Get returns a single notice for a name and a language
	Get(ctx context.Context, name PinnedNoticeName, language string) (*Notice, error)
}

// NoticesService is the low level store to manage single notices
type NoticesService interface {
	// GetByID returns the page for that ID or an error
	GetByID(context.Context, int64) (Notice, error)

	// Save updates the passed page or creates it if it's ID is zero
	Save(context.Context, *Notice) error

	// RemoveID removes the page for that ID.
	RemoveID(context.Context, int64) error
}

// for tests we use generated mocks from these interfaces created with https://github.com/maxbrunsfeld/counterfeiter

//go:generate counterfeiter -o mockdb/auth.go . AuthWithSSBService

//go:generate counterfeiter -o mockdb/auth_fallback.go . AuthFallbackService

//go:generate counterfeiter -o mockdb/allow.go . AllowListService

//go:generate counterfeiter -o mockdb/alias.go . AliasService

//go:generate counterfeiter -o mockdb/invite.go . InviteService

//go:generate counterfeiter -o mockdb/fixed_pages.go . PinnedNoticesService

//go:generate counterfeiter -o mockdb/pages.go . NoticesService

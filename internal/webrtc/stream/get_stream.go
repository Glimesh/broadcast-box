package stream

import (
	"context"
	"strings"

	"github.com/glimesh/broadcast-box/internal/server/authorization"
)

// TODO:
// When 2 different machines start a session, the latter will overwrite the session id
// and take over the session. To allow for a smooth multiuser stream against one streamKey,
// make the session Id into a list of session Ids that are maintained as sessions are added and removed.
// Doing so will also allow for layer changes to be connected, so that changing the video feed triggers the
// corresponding audio feed as well, if at another session
func GetStream(whipSessions map[string]*WhipSession, profile authorization.Profile, whipSessionId string) (*WhipSession, error) {
	stream, ok := whipSessions[profile.StreamKey]

	if !ok {
		whipActiveContext, whipActiveContextCancel := context.WithCancel(context.Background())

		stream = &WhipSession{
			StreamKey:           strings.ReplaceAll(profile.StreamKey, " ", ""),
			IsPublic:            profile.IsPublic,
			MOTD:                profile.MOTD,
			SessionId:           whipSessionId,
			ActiveContext:       whipActiveContext,
			ActiveContextCancel: whipActiveContextCancel,
			PliChan:             make(chan any, 250),

			AudioTracks: []*AudioTrack{},
			VideoTracks: []*VideoTrack{},

			WhepSessions: map[string]*WhepSession{},
		}

		whipSessions[profile.StreamKey] = stream
	}

	if whipSessionId != "" {
		stream.SessionId = whipSessionId
		stream.HasHost.Store(true)
	}

	return stream, nil
}

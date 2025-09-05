package session

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/glimesh/broadcast-box/internal/server/authorization"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v4"
)

// TODO:
// When 2 different machines start a session, the latter will overwrite the session id
// and take over the session. To allow for a smooth multiuser stream against one streamKey,
// make the session Id into a list of session Ids that are maintained as sessions are added and removed.
// Doing so will also allow for layer changes to be connected, so that changing the video feed triggers the
// corresponding audio feed as well, if at another session
func GetStream(profile authorization.PublicProfile, whipSessionId string) (*WhipSession, error) {
	WhipSessionsLock.Lock()
	stream, ok := WhipSessions[profile.StreamKey]

	if !ok {
		whipActiveContext, whipActiveContextCancel := context.WithCancel(context.Background())

		stream = &WhipSession{
			SessionId: whipSessionId,

			StreamKey: strings.ReplaceAll(profile.StreamKey, " ", ""),
			IsPublic:  profile.IsPublic,
			MOTD:      profile.MOTD,

			ActiveContext:       whipActiveContext,
			ActiveContextCancel: whipActiveContextCancel,
			PliChan:             make(chan any, 250),
			OnOnlineChan:        make(chan bool, 50),
			OnTrackChan:         make(chan struct{}, 50),
			SSEChan:             make(chan any, 50),

			AudioTracks: []*AudioTrack{},
			VideoTracks: []*VideoTrack{},

			WhepSessions: map[string]*WhepSession{},
		}

		WhipSessions[profile.StreamKey] = stream
	}

	if whipSessionId != "" {
		stream.SessionId = whipSessionId
		stream.HasHost.Store(true)
		stream.MOTD = profile.MOTD
	}

	WhipSessionsLock.Unlock()
	return stream, nil
}

func StartWhipSessionLoop(stream *WhipSession) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		// All WHIP and WHEP sessions has left, close down
		case <-stream.ActiveContext.Done():
			ticker.Stop()
			return

		// Announce new layers available
		case <-stream.OnTrackChan:
			// Lock, copy session data, then unlock
			WhipSessionsLock.Lock()
			stream.WhepSessionsLock.RLock()

			whepSessionsCopy := make([]*WhepSession, 0, len(stream.WhepSessions))
			for _, whep := range stream.WhepSessions {
				whepSessionsCopy = append(whepSessionsCopy, whep)
			}

			stream.WhepSessionsLock.RUnlock()
			WhipSessionsLock.Unlock()

			// Generate layer info outside lock
			currentLayers := GetAvailableLayersJsonString(stream)

			// Send to each WHEP session
			for _, whep := range whepSessionsCopy {
				select {
				case whep.SSEChannel <- currentLayers:
				default:
					log.Println("WHIP.Loop: OnTrackChange - Channel full, skipping update")
				}
			}

		// Send status every 5 seconds
		case <-ticker.C:
			// Lock, copy session data, then unlock
			stream.WhepSessionsLock.RLock()
			whepSessionsCopy := make([]*WhepSession, 0, len(stream.WhepSessions))
			for _, whep := range stream.WhepSessions {
				whepSessionsCopy = append(whepSessionsCopy, whep)
			}
			stream.WhepSessionsLock.RUnlock()

			// Generate status
			currentStatus := GetSessionStatsJsonString(stream)

			// Send to WHIP session
			stream.SSEChan <- currentStatus

			// Send to each WHEP session
			for _, whep := range whepSessionsCopy {
				select {
				case whep.SSEChannel <- currentStatus:
				default:
					log.Println("WHIP.Loop: Status update skipped for session due to full channel")
				}
			}
		}
	}
}

func RtcpPacketReaderLoop(stream *WhipSession, rtpSender *webrtc.RTPSender) {
	for {
		rtcpPackets, _, rtcpErr := rtpSender.ReadRTCP()
		if rtcpErr != nil {
			return
		}

		for _, r := range rtcpPackets {
			if _, isPli := r.(*rtcp.PictureLossIndication); isPli {
				select {
				case stream.PliChan <- true:
				default:
				}
			}
		}
	}
}

// Get highest prioritized audio track in the whip session
// This only works if the priority has been set.
// Currently this is only supported when being set through the simulcast
// property in the offer made by the whip connection
func (session *WhipSession) GetHighestPrioritizedAudioTrack() string {
	if len(session.AudioTracks) != 0 {
		highestPriorityAudioTrack := session.AudioTracks[0]
		for _, trackPriority := range session.AudioTracks[1:] {
			if trackPriority.Priority < highestPriorityAudioTrack.Priority {
				highestPriorityAudioTrack = trackPriority
			}
		}

		return highestPriorityAudioTrack.Rid
	}

	return ""
}

// Get highest prioritized video track in the whip session
// This only works if the priority has been set.
// Currently this is only supported when being set through the simulcast
// property in the offer made by the whip connection
func (session *WhipSession) GetHighestPrioritizedVideoTrack() string {
	if len(session.VideoTracks) != 0 {
		highestPriorityVideoTrack := session.VideoTracks[0]
		for _, trackPriority := range session.VideoTracks[1:] {
			if trackPriority.Priority < highestPriorityVideoTrack.Priority {
				highestPriorityVideoTrack = trackPriority
			}
		}

		return highestPriorityVideoTrack.Rid
	}

	return ""
}

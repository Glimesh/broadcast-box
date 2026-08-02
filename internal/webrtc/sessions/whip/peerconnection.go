package whip

import (
	"log/slog"

	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v4"
)

func (w *WHIPSession) SetOnClosed(onClosed func()) {
	w.onClosed = onClosed
}

func (w *WHIPSession) notifyClosed() {
	w.closeOnce.Do(func() {
		if w.onClosed != nil {
			w.onClosed()
		}
	})
}

func (w *WHIPSession) AddPeerConnection(peerConnection *webrtc.PeerConnection, streamKey string) {
	slog.Info("WHIPSession.AddPeerConnection")

	w.PeerConnectionLock.Lock()
	existingPeerConnection := w.PeerConnection
	w.PeerConnection = peerConnection
	w.PeerConnectionLock.Unlock()

	if existingPeerConnection != nil && existingPeerConnection != peerConnection {
		slog.Info("WHIPSession.AddPeerConnection: Replacing existing peerconnection")
		if err := existingPeerConnection.GracefulClose(); err != nil {
			slog.Error("WHIPSession.AddPeerConnection.Close.Error", "err", err)
		}
	}

	w.registerWHIPHandlers(peerConnection, streamKey)
}

func (w *WHIPSession) RemovePeerConnection() {
	slog.Info("WHIPSession.RemovePeerConnection", "id", w.ID)

	w.PeerConnectionLock.Lock()
	peerConnection := w.PeerConnection
	w.PeerConnection = nil
	w.PeerConnectionLock.Unlock()

	if peerConnection == nil {
		return
	}

	if err := peerConnection.Close(); err != nil {
		slog.Error("WHIPSession.RemovePeerConnection.Error", "err", err)
	}

	slog.Info("WHIPSession.RemovePeerConnection.Completed", "id", w.ID)
}

func (w *WHIPSession) SendPLI() {
	w.PeerConnectionLock.RLock()
	peerConnection := w.PeerConnection
	w.PeerConnectionLock.RUnlock()
	if peerConnection == nil {
		return
	}

	packets := w.getPLIPackets()
	if len(packets) == 0 {
		return
	}

	if err := peerConnection.WriteRTCP(packets); err != nil {
		slog.Error("WHIPSession.SendPLI.WriteRTCP.Error", "err", err)
	}
}

func (w *WHIPSession) getPLIPackets() []rtcp.Packet {
	w.TracksLock.RLock()
	defer w.TracksLock.RUnlock()

	packets := make([]rtcp.Packet, 0, len(w.VideoTracks))
	for _, track := range w.VideoTracks {
		if mediaSSRC := track.MediaSSRC.Load(); mediaSSRC != 0 {
			packets = append(packets, &rtcp.PictureLossIndication{
				MediaSSRC: mediaSSRC,
			})
		}
	}

	return packets
}

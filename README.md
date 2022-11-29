# Broadcast Box
Broadcast Box lets you broadcast to others in sub-second time. It was designed
to be simple to use and easily modifiable. We wrote Broadcast Box to show off some
of the cutting edge tech that is coming to the broadcast space.

### Subsecond Latency
Broadcast Box uses WebRTC for broadcast and playback. By using WebRTC instead of
RTMP and HLS you get the fastest experience possible.

### Peer-to-Peer (if you need it)
With Broadcast Box you can serve your video without a public IP or forwarding ports!
Run Broadcast Box on the same machine that you are running OBS, and share your
video with the world! WebRTC comes with P2P technology, so users can broadcast
and playback video without paying for dedicated servers.

### Latest in Video Compression
With WebRTC you get access to the latest in video codecs. With AV1 you can send
the same video quality with a (INSERT MARKETING NUMBERS) reduction in bandwidth.

### Broadcast all angles
WebRTC allows you to upload multiple video streams in the same session. Now you can
broadcast multiple camera angles, or share interactive video experiences in real time!

### Broadcasters provide transcodes
Transcodes are necessary if you want to provide a good experience to all your users.
Generating them is prohibitively though. WebRTC provides a solution. With WebRTC
users can upolad the same video at different quality levels. This
keeps things cheap for the server operator and you still can provide the same
experience.

# Running
### Docker
### Directly
### P2P Bridge


# Design
Broadcast Box has a Go backend and React frontend. The backend exposes four endpoints currently.

* `/api/status` - Has the server gone through initial configuration yet
* `/api/configure` - Configure things like server stream key and admin password
* `/api/whip` - Start a WHIP Session. WHIP broadcasts video via WebRTC.
* `/api/whep` - Start a WHEP Session. WHEP is video playback via WebRTC.

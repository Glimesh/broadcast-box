# Broadcast Box

[![License][license-image]][license-url]
[![Discord][discord-image]][discord-invite-url]

- [What is Broadcast Box](#what-is-broadcast-box)
- [Using](#using)
  - [Broadcasting](#broadcasting)
  - [Broadcasting (GStreamer, CLI)](#broadcasting-gstreamer-cli)
  - [Playback](#playback)
- [Getting Started](#getting-started)
  - [Configuring](#configuring)
  - [Building From Source](#building-from-source)
  - [Frontend](#frontend)
  - [Backend](#backend)
  - [Docker](#docker)
  - [Docker Compose](#docker-compose)
  - [Environment variables](#environment-variables)
  - [Network Test on Start](#network-test-on-start)
- [Design](#design)

## What is Broadcast Box

Broadcast Box lets you broadcast to others in sub-second time. It was designed
to be simple to use and easily modifiable. We wrote Broadcast Box to show off some
of the cutting edge tech that is coming to the broadcast space.

Want to contribute to the development of Broadcast Box? See [Contributing](./CONTRIBUTING.md).

### Sub-second Latency

Broadcast Box uses WebRTC for broadcast and playback. By using WebRTC instead of
RTMP and HLS you get the fastest experience possible.

### Latest in Video Compression

With WebRTC you get access to the latest in video codecs. With AV1 you can send
the same video quality with a [50%][av1-practical-use-case] reduction in bandwidth required.

[av1-practical-use-case]: https://engineering.fb.com/2018/04/10/video-engineering/av1-beats-x264-and-libvpx-vp9-in-practical-use-case/

### Broadcast all angles

WebRTC allows you to upload multiple video streams in the same session. Now you can
broadcast multiple camera angles, or share interactive video experiences in real time!

### Broadcasters provide transcodes

Transcodes are necessary if you want to provide a good experience to all your users.
Generating them is prohibitively expensive though, WebRTC provides a solution. With WebRTC
users can upload the same video at different quality levels. This
keeps things cheap for the server operator and you still can provide the same
experience.

### Broadcasting for all

WebRTC means anyone can be a broadcaster. With Broadcast Box you could use broadcast software like OBS.
However, another option is publishing directly from your browser! Users just getting started with streaming
don't need to worry about bitrates, codecs anymore. With one press of a button you can go live right from
your browser with Broadcast Box. This makes live-streaming accessible to an entirely new audience.

### Peer-to-Peer (if you need it)

With Broadcast Box you can serve your video without a public IP or forwarding ports!

Run Broadcast Box on the same machine that you are running OBS, and share your
video with the world! WebRTC comes with P2P technology, so users can broadcast
and playback video without paying for dedicated servers. To start the connection users will
need to be able to connect to the HTTP server. After they have negotiated the session then
NAT traversal begins.

You could also use P2P to pull other broadcasters into your stream. No special configuration
or servers required anymore to get sub-second co-streams.

Broadcast Box acts as a [SFU][applied-webrtc-article]. This means that
every client connects to Broadcast Box. No direct connection is established between broadcasters/viewers.
If you want a direct connection between OBS and your browser see [OBS2Browser][obs-2-browser-repo].

[applied-webrtc-article]: https://webrtcforthecurious.com/docs/08-applied-webrtc/#selective-forwarding-unit
[obs-2-browser-repo]: https://github.com/Sean-Der/OBS2Browser

## Using

To use Broadcast Box you don't even have to run it locally! A instance of Broadcast Box
is hosted at [b.siobud.com](https://b.siobud.com). If you wish to run it locally skip to [Getting Started](#getting-started)

### Broadcasting

To use Broadcast Box with OBS you must set your output to WebRTC and set a proper URL + Stream Key.
You may use any Stream Key you like. The same stream key is used for broadcasting and playback.

Go to `Settings -> Stream` and set the following values.

- Service: WHIP
- Server: <https://b.siobud.com/api/whip>
- StreamKey: (Any Stream Key you like)

Your settings page should look like this:

![OBS Stream settings example](./.github/img/streamSettings.png)

OBS by default will have ~2 seconds of latency. If you want sub-second latency you can configure
this in `Settings -> Output`. Set your encoder to `x264` and set tune to `zerolatency`. Your Output
page will look like this.

![OBS Output settings example](./.github/img/outputPage.png)

When you are ready to broadcast press `Start Streaming` and now time to watch!

### Broadcasting (GStreamer, CLI)

See the example script [here](examples/gstreamer-broadcast.nu).

Can broadcast gstreamer's test sources, or pulsesrc+v4l2src

Expects `gstreamer-1.0`, with `good,bad,ugly` plugins and `gst-plugins-rs`

Use of example scripts:

```shell
# testsrcs
./examples/gstreamer-broadcast.nu http://localhost:8080/api/whip testStream1
# v4l2src
./examples/gstreamer-broadcast.nu http://localhost:8080/api/whip testStream1 v4l2
```

### Playback

If you are broadcasting to the Stream Key `StreamTest` your video will be available at <https://b.siobud.com/StreamTest>.

You can also go to the home page and enter `StreamTest`. The following is a screenshot of OBS broadcasting and
the latency of 120 milliseconds observed.

![Example have potential latency](./.github/img/broadcastView.png)

## Getting Started

Broadcast Box is made up of three parts. The server is written in Go and is in charge of ingesting and broadcasting WebRTC. The authentication-backend is written in go and utilizes pocketbase to abstract most api endpoints.The frontend is in react and connects to the Go backend. The Go server can be used to serve the HTML/CSS/JS directly. Use the following instructions to build from source or utilize [Docker](#docker) / [Docker Compose](#docker-compose).

### Configuring

Configurations can be made in [.env.production](./.env.production), although the defaults should get things going.

### Building From Source

#### Frontend

React dependencies are installed by running `npm install` in the `web` directory and `npm run build` will build the frontend.

If everything is successful, you should see the following:

```console
> broadcast-box@0.1.0 build
> dotenv -e ../.env.production react-scripts build

Creating an optimized production build...
Compiled successfully.

File sizes after gzip:

  53.51 kB  build/static/js/main.12067218.js
  2.27 kB   build/static/css/main.8738ee38.css
...
```

#### Backend

Go dependencies are automatically installed.

To run the Go server, run `go run .` in the root of this project, you should see the following:

```console
2022/12/11 16:02:14 Loading `.env.production`
2022/12/11 16:02:14 Running HTTP Server at `:8080`
```

To use Broadcast Box navigate to: `http://<YOUR_IP>:8080`. In your broadcast tool of choice, you will broadcast to `http://<YOUR_IP>:8080/api/whip`.

#### Authetication-backend
The authentication backend can be started by going into the `authentication-backend` folder and run:
```shell
go run main.go serve 
```

### Docker

A Docker image is also provided to make it easier to run locally and in production. The arguments you run the Dockerfile with depending on
if you are using it locally or a server.

If you want to run locally execute `docker run -e UDP_MUX_PORT=8080 -e NAT_1_TO_1_IP=127.0.0.1 -p 8080:8080 -p 8080:8080/udp seaduboi/broadcast-box`.
This will make broadcast-box available on `http://localhost:8080`. The UDPMux is needed because Docker on macOS/Windows runs inside a NAT.

If you are running on AWS (or other cloud providers) execute. `docker run --net=host -e INCLUDE_PUBLIC_IP_IN_NAT_1_TO_1_IP=yes seaduboi/broadcast-box`
broadcast-box needs to be run in net=host mode. broadcast-box listens on random UDP ports to establish sessions.

### Docker Compose

A Docker Compose is included that uses LetsEncrypt for automated HTTPS. It also includes Watchtower so your instance of Broadcast Box
will be automatically updated every night. If you are running on a VPS/Cloud server this is the quickest/easiest way to get started.

```console
export URL=my-server.com
docker-compose up -d
```
## URL Parameters

The frontend can be configured by passing these URL Parameters.

- `cinemaMode=true` - Forces the player into cinema mode by adding to end of URL like https://b.siobud.com/myStream?cinemaMode=true

## Environment Variables

The backend can be configured with the following environment variables.

- `DISABLE_STATUS` - Disable the status API
- `DISABLE_FRONTEND` - Disable the serving of frontend. Only REST APIs + WebRTC is enabled.
- `HTTP_ADDRESS` - HTTP Server Address
- `NETWORK_TEST_ON_START` - When "true" on startup Broadcast Box will check network connectivity

- `ENABLE_HTTP_REDIRECT` - HTTP traffic will be redirect to HTTPS
- `SSL_CERT` - Path to SSL certificate if using Broadcast Box's HTTP Server
- `SSL_KEY` - Path to SSL key if using Broadcast Box's HTTP Server

- `NAT_1_TO_1_IP` - Announce IPs that don't belong to local machine (like Public IP). delineated by '|'
- `INCLUDE_PUBLIC_IP_IN_NAT_1_TO_1_IP` - Like `NAT_1_TO_1_IP` but autoconfigured
- `INTERFACE_FILTER` - Only use a certain interface for UDP traffic
- `NAT_ICE_CANDIDATE_TYPE` - By default setting a NAT_1_TO_1_IP overrides. Set this to `srflx` to instead append IPs
- `STUN_SERVERS` - List of STUN servers delineated by '|'. Useful if Broadcast Box is running behind a NAT
- `INCLUDE_LOOPBACK_CANDIDATE` - Also listen for WebRTC traffic on loopback, disabled by default

- `UDP_MUX_PORT_WHEP` - Like `UDP_MUX_PORT` but only for WHEP traffic
- `UDP_MUX_PORT_WHIP` - Like `UDP_MUX_PORT` but only for WHIP traffic
- `UDP_MUX_PORT` - Serve all UDP traffic via one port. By default Broadcast Box listens on a random port

- `TCP_MUX_ADDRESS` - If you wish to make WebRTC traffic available via TCP.
- `TCP_MUX_FORCE` - If you wish to make WebRTC traffic only available via TCP.

- `APPEND_CANDIDATE` - Append candidates to Offer that ICE Agent did not generate. Worse version of `NAT_1_TO_1_IP`

- `DEBUG_PRINT_OFFER` - Print WebRTC Offers from client to Broadcast Box. Debug things like accepted codecs.
- `DEBUG_PRINT_ANSWER` - Print WebRTC Answers from Broadcast Box to Browser. Debug things like IP/Ports returned to client.

## Network Test on Start

When running in Docker Broadcast Box runs a network tests on startup. This tests that WebRTC traffic can be established
against your server. If you server is misconfigured Broadcast Box will not start.

If the network test is enabled this will be printed on startup

```console
NETWORK_TEST_ON_START is enabled. If the test fails Broadcast Box will exit.
See the README.md for how to debug or disable NETWORK_TEST_ON_START
```

If the test passed you will see

```console
Network Test passed.
Have fun using Broadcast Box
```

If the test failed you will see the following. The middle sentence will change depending on the error.

```console
Network Test failed.
Network Test client reported nothing in 30 seconds
Please see the README and join Discord for help
```

[Join the Discord][discord-invite-url] and we are ready to help! To debug check the following.

- Have you allowed UDP traffic?
- Do you have any restrictions on ports?
- Is your server publicly accessible?

If you wish to disable the test set the environment variable `NETWORK_TEST_ON_START` to false.

## Design

The backend exposes three endpoints (the status page is optional, if hosting locally).

- `/api/whip` - Start a WHIP Session. WHIP broadcasts video via WebRTC.
- `/api/whep` - Start a WHEP Session. WHEP is video playback via WebRTC.
- `/api/status` - Status of the all active WHIP streams

[license-image]: https://img.shields.io/badge/License-MIT-yellow.svg
[license-url]: https://opensource.org/licenses/MIT
[discord-image]: https://img.shields.io/discord/1162823780708651018?logo=discord
[discord-invite-url]: https://discord.gg/An5jjhNUE3

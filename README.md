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
Broadcast Box is made up of two parts. The server is written in Go and is in charge
of ingesting and broadcasting WebRTC. The frontend is in react and connects to the Go
backend.

While developing `webpack-dev-server` is used for the frontend. In production the Go server
can be used to serve the HTML/CSS/JS directly. These are the instructions on how to run all
these parts.

### Installing Dependencies
Go dependencies are automatically installed.

react dependencies are installed by running `npm install` in the `web` directory.

### Configuring
Both projects use `.env` files for configuration. For development `.env.development` is used
and in production `.env.production` is used.

For Go setting `APP_ENV` will cause `.env.production` to be loaded.
Otherwise `.env.development` is used.

For react App the dev server uses `.env.development` and `npm run build`
uses `.env.production`

### Local Development
For local development you will run the Go server and webpack directly.

To run the Go server run `go run .` in the root of this project. You will see the logs
like the following.

```
2022/12/11 15:22:47 Loading `.env.development`
2022/12/11 15:22:47 Running HTTP Server at `:8080`
```

To run the web front open the `web` folder and execute `npm start` if that runs successfully you will
be greeted with.

```
Compiled successfully!

You can now view broadcast-box in the browser.

  Local:            http://localhost:3000
  On Your Network:  http://192.168.1.57:3000

Note that the development build is not optimized.
To create a production build, use npm run build.

webpack compiled successfully
```

To use Broadcast Box you will open `http://localhost:3000` in your browser. In your broadcast tool of choice
you will broadcast to `http://localhost:8080/api/whip`.

### Production
For production usage Go will server the frontend and backend.

To run the Go server run `APP_ENV=production go run .` in the root of this project. You will see the logs
like the following.

```
2022/12/11 16:02:14 Loading `.env.production`
2022/12/11 16:02:14 Running HTTP Server at `:8080`
```

If `APP_ENV` was set properly `.env.production` will be loaded.

To build the frontend execute `npm run build` in the `web` directory. You will get the following output.

```
> broadcast-box@0.1.0 build
> dotenv -e ../.env.production react-scripts build

Creating an optimized production build...
Compiled successfully.

File sizes after gzip:

  53.51 kB  build/static/js/main.12067218.js
  2.27 kB   build/static/css/main.8738ee38.css
```

To use Broadcast Box you will open `http://localhost:8080` in your browser. In your broadcast tool of choice
you will broadcast to `http://localhost:8080/api/whip`.

### Docker

# Design
Broadcast Box has a Go backend and React frontend. The backend exposes two endpoints.

* `/api/whip` - Start a WHIP Session. WHIP broadcasts video via WebRTC.
* `/api/whep` - Start a WHEP Session. WHEP is video playback via WebRTC.

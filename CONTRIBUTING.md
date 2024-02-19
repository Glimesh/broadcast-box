# Contributing to Broadcast Box

Contributing to broadcast box is greatly appreciated. We are happy to give guidance, answer questions and review PRs.

## Getting Started

Broadcast Box is made up of two parts. The server is written in Go and is in charge
of ingesting and broadcasting WebRTC. The frontend is in react and connects to the Go
backend.

### Configuring

Configurations can be made in [.env.development](./.env.development).

### Frontend

React dependencies are installed by running `npm install` in the `web` directory, then run `npm start` to serve the frontend. You should see the following:

```console
Compiled successfully!

You can now view broadcast-box in the browser.

  Local:            http://localhost:3000
  On Your Network:  http://192.168.1.57:3000

Note that the development build is not optimized.
To create a production build, use npm run build.

webpack compiled successfully
```

### Backend

Go dependencies are automatically installed.

To run the Go server run `APP_ENV=development go run .` in the root of this project. You will see the logs
like the following.

```console
2022/12/11 15:22:47 Loading `.env.development`
2022/12/11 15:22:47 Running HTTP Server at `:8080`
```

To use Broadcast Box navigate to: `http://localhost:3000`. In your broadcast tool of choice, you will broadcast to `http://localhost:8080/api/whip`.

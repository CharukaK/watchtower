# Watchtower

Watchtower is a local web dashboard for Hikvision-compatible DVRs and cameras.
It reads device and channel details through ISAPI, registers enabled RTSP
streams with MediaMTX, and plays them in the browser over WebRTC.

![Watchtower dashboard with synthetic camera data](docs/watchtower-dashboard.png)

## What works

- Live camera grid with two- and three-column layouts
- On-demand RTSP stream registration in MediaMTX
- Browser playback over WebRTC
- Camera snapshots and stream details
- Device, channel, and MediaMTX status
- Channel capability inspection

The recent-events and recording-replay screens are placeholders. They are not connected to the DVR yet.

## Requirements

The Docker setup requires Docker Compose and a Hikvision-compatible device with
ISAPI enabled. Manual development also requires Go 1.27.1, Node.js, npm, and
[MediaMTX](https://github.com/bluenviron/mediamtx).

The machine running Watchtower must be able to reach the device. Browsers reach
MediaMTX's WebRTC HTTP endpoints through Watchtower's `/media/` proxy, but must
still be able to reach MediaMTX's advertised WebRTC media address.

## Quick start with Docker

Copy the environment template and fill in the device address, credentials, and
this machine's LAN address:

```sh
cp .env.example .env
$EDITOR .env
docker compose up --build -d
```

Open <http://localhost:8080>. The default configuration binds Watchtower to
loopback. To use it from another trusted LAN device, set
`WATCHTOWER_BIND_ADDRESS` to the server's LAN IP and open that address in the
browser. Allow the configured TCP port and UDP port `8189` through the host
firewall.

View logs or stop the services with:

```sh
docker compose logs -f
docker compose down
```

## Manual development setup

Install the frontend tools:

```sh
npm ci
```

Set `MTX_WEBRTCADDITIONALHOSTS` to the LAN address of the machine running
MediaMTX. Use the same control API password for MediaMTX and Watchtower:

```sh
MTX_WEBRTCADDITIONALHOSTS=192.168.1.10 \
MTX_AUTHINTERNALUSERS_1_PASS=change-me-too \
mediamtx mediamtx.yml
```

In another terminal, start Watchtower:

```sh
HOST_URL=http://dvr.local \
USER=admin \
PASSWORD=change-me \
MEDIAMTX_HOST=http://127.0.0.1:9997 \
MEDIAMTX_USER=user \
MEDIAMTX_PASSWORD=change-me-too \
make dev
```

Open <http://localhost:7331>. Air proxies the application through port `7331`
and reloads it when Go, templ, CSS, or JavaScript files change. To connect from
another trusted LAN device, use the server's LAN IP and allow TCP port `7331`
and UDP port `8189` through the host firewall.

## Configuration

| Variable | Required | Description |
| --- | --- | --- |
| `HOST_URL` | Yes | Device origin, such as `http://dvr.local`. Do not include `/ISAPI` or credentials. |
| `USER` | Yes | Device username. |
| `PASSWORD` | Yes | Device password. |
| `MEDIAMTX_HOST` | Yes | MediaMTX API origin, usually `http://127.0.0.1:9997`. |
| `MEDIAMTX_USER` | Yes | MediaMTX API username. |
| `MEDIAMTX_PASSWORD` | Yes | MediaMTX API password. |
| `PORT` | No | Watchtower HTTP port. Defaults to `8080`. |

The Compose setup maps `DVR_USER` and `DVR_PASSWORD` to the application's
`USER` and `PASSWORD` variables. It also uses `WEBRTC_ADDITIONAL_HOST`,
`WATCHTOWER_BIND_ADDRESS`, and `WATCHTOWER_PORT`; see `.env.example`.

Watchtower adds one on-demand MediaMTX path per enabled channel, named
`channel-<channel ID>`. Existing paths with those names are left unchanged.

For a container deployment, use the MediaMTX service name for the private API:

```sh
MEDIAMTX_HOST=http://mediamtx:9997
```

Watchtower derives the internal WebRTC HTTP target as `http://mediamtx:8889`.
Publish Watchtower's HTTP port and MediaMTX's UDP port `8189`; the MediaMTX API
and HTTP port can remain on the container network. Set `webrtcAdditionalHosts`
to an address that browsers outside the container network can reach.

## Build and test

```sh
make build
make test
```

The production binary is written to `build/watchtower`. Run it with the same
environment variables shown above, then open port `8080` or the value set in
`PORT`.

Other useful commands:

```sh
make generate  # regenerate templ files
make css       # rebuild the Tailwind CSS bundle
make clean     # remove generated build output
```

## Security notes

- Watchtower does not have login protection. Keep it on a trusted network or
put an authenticated reverse proxy in front of it.
- The Compose setup exposes Watchtower only on loopback unless you change
`WATCHTOWER_BIND_ADDRESS`.
- Media playback is anonymous inside the trusted network. The MediaMTX control
API is not published by the Compose setup and uses the password from `.env`.
- Device credentials are placed in MediaMTX's in-memory RTSP path configuration
so MediaMTX can connect to the device.
- Do not put credentials in `HOST_URL` or `MEDIAMTX_HOST`, and do not commit
`.env`.

## License

Watchtower is available under the [MIT License](LICENSE).

package web

import "embed"

// Static contains the browser assets bundled into the executable.
//
//go:embed static/app.css static/datastar.js static/player.js
var Static embed.FS

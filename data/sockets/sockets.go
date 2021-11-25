package sockets

import (
	"regexp"
)

var reURL = regexp.MustCompile(`(wss?://[a-zA-Z0-9-_.:]+/ws)/([a-zA-Z0-9]{6,})$`)

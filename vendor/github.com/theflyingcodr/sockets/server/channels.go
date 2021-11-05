package server

type channel struct {
	id    string
	conns map[string]*connection
}

func newChannel(id string) *channel {
	r := &channel{
		id:    id,
		conns: make(map[string]*connection),
	}
	return r
}

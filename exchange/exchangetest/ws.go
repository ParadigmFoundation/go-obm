package exchangetest

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

type WsFn func(*testing.T, *websocket.Conn)

type WS struct {
	t  *testing.T
	fn WsFn
}

func NewWS(t *testing.T, fn WsFn) *WS {
	return &WS{
		t:  t,
		fn: fn,
	}
}

func (ws *WS) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{}
	conn, err := upgrader.Upgrade(w, r, nil)

	require.NoError(ws.t, err)
	if fn := ws.fn; fn != nil {
		fn(ws.t, conn)
	}
}

func (ws *WS) Start() (string, func()) {
	srv := httptest.NewServer(ws)
	url := strings.Replace(srv.URL, "http", "ws", 1)
	return url, srv.Close
}

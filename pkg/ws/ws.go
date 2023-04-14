package ws

// import (
// 	"net/http"

// 	"github.com/questx-lab/backend/internal/repository"
// )

// type WebSocket interface {
// 	Serve(w http.ResponseWriter, r *http.Request)
// }

// func NewWebSocket(mux *http.ServeMux, roomRepo repository.RoomRepository) WebSocket {
// 	hub := newHub()

// 	ws := &webSocket{
// 		Hub:      hub,
// 		roomRepo: roomRepo,
// 	}

// 	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
// 		serveWs(hub, w, r)
// 	})
// 	return &webSocket{}
// }

// serveWs handles websocket requests from the peer.

package guiapi

import (
	"context"
	"log"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

func (h *Handler) websocketHandler(c *Context) {
	conn, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
		Subprotocols: []string{"guiapi"},
	})
	if err != nil {
		log.Println("websocket accept error:", err)
		return
	}
	defer conn.Close(websocket.StatusInternalError, "the sky is falling")

	if conn.Subprotocol() != "guiapi" {
		log.Printf("websocket accept error: invalid subprotocol %q", conn.Subprotocol())
		return
	}
	log.Println("websocket accepted")

	ctx, cancel := context.WithTimeout(c.Request.Context(), time.Hour)
	defer cancel()

	var v interface{}
	err = wsjson.Read(ctx, conn, &v)
	if err != nil {
		log.Println("websocket read error:", err)
		return
	}

	log.Printf("websocket received: %v", v)

	for {
		type Message struct {
			Message string
			Time    time.Time
		}
		err := wsjson.Write(ctx, conn, Message{"hello from server", time.Now()})
		if err != nil {
			log.Println("websocket write error:", err)
			break
		}
		time.Sleep(2 * time.Second)
	}

	conn.Close(websocket.StatusNormalClosure, "")
}

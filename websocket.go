package guiapi

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"nhooyr.io/websocket"
)

type StreamFunc func(ctx context.Context, args json.RawMessage, res chan<- *Response) error

type websocketMessage struct {
	Name string          `json:"name"`
	Args json.RawMessage `json:"args"`
}

func (s *Server) websocketHandler(c *PageCtx) {
	conn, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
		Subprotocols: []string{"guiapi"},
	})
	if err != nil {
		log.Println("websocket accept error:", err)
		return
	}
	defer conn.Close(websocket.StatusInternalError, "exit")

	if conn.Subprotocol() != "guiapi" {
		log.Printf("websocket accept error: invalid subprotocol %q", conn.Subprotocol())
		return
	}

	ctx := c.Request.Context()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ch := make(chan *Response, 1)
	defer close(ch)

	id := time.Now().UnixMilli() % 1000

	go func() {
		defer log.Println("exit websocket writer", id)
		for {
			select {
			case <-ctx.Done():
				err := conn.Close(websocket.StatusNormalClosure, "done")
				if err != nil {
					log.Println("websocket close error:", err)
					return
				}
				return
			case resp, ok := <-ch:
				if !ok {
					log.Println("websocket writer not ok", id)
					return
				}
				buf, err := json.Marshal(resp)
				if err != nil {
					log.Println("json marshal error:", err)
					return
				}

				err = conn.Write(ctx, websocket.MessageText, buf)
				if err != nil {
					log.Println("websocket write error:", err)
					return
				}
			}
		}
	}()

	messages := make(chan []byte)
	defer close(messages)

	go func() {
		defer log.Println("exit websocket reader", id)
		for {
			msgType, buf, err := conn.Read(ctx)
			if err != nil {
				log.Println("websocket read error:", err)
				cancel()
				return
			}
			if msgType != websocket.MessageText {
				log.Println("websocket read error: invalid message type", msgType)
				return
			}
			select {
			case messages <- buf:
			case <-ctx.Done():
				log.Println("websocket reader blocked", id)
			}
		}
	}()

	var previousCancel context.CancelFunc
	for {
		defer log.Println("exit websocket router", id)
		select {
		case <-ctx.Done():
			return
		case buf, ok := <-messages:
			if previousCancel != nil {
				previousCancel()
			}
			if !ok {
				log.Println("websocket router not ok", id)
				return
			}
			var msg websocketMessage
			err := json.Unmarshal(buf, &msg)
			if err != nil {
				log.Println("json unmarshal error:", err)
				cancel()
				break
			}
			subCtx, subCancel := context.WithCancel(ctx)
			previousCancel = subCancel
			go func() {
				fn := s.streams[msg.Name]
				if fn == nil {
					log.Println("StreamRouter error: unknown stream", msg.Name)
					cancel()
					return
				}
				err := fn(subCtx, msg.Args, ch)
				if err != nil {
					log.Println("StreamRouter error:", err)
					cancel()
				}
			}()
		}
	}
}

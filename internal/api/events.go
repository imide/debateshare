package api

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func (s *Server) roomEvents(c *gin.Context) {
	room, ok := s.liveRoom(c)
	if !ok {
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	ch, unsub := s.hub.Subscribe(room.Code)
	defer unsub()

	heartbeat := time.NewTicker(25 * time.Second)
	defer heartbeat.Stop()

	_, err := fmt.Fprintf(c.Writer, ": connected\n\n")
	if err != nil {
		log.Err(err).Msg("failed to write to client")
		return
	}
	c.Writer.Flush()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case <-heartbeat.C:
			fmt.Fprintf(c.Writer, ": ping\n\n")
			c.Writer.Flush()
		case ev, open := <-ch:
			if !open {
				return
			}
			fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", ev.Name, ev.Data)
			c.Writer.Flush()
		}
	}
}

package api

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func fail(c *gin.Context, status int, msg string) {
	c.AbortWithStatusJSON(status, gin.H{"error": msg})
	log.Debug().Int("status", status).Msg(msg)
}

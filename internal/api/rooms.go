package api

import (
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/imide/debateshare/internal/models"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// emailRe (fyi AI generated this, i hate regex lol)
var emailRe = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

type roomRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (s *Server) createRoom(c *gin.Context) {
	var req roomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid json body")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if len(req.Name) > 120 {
		fail(c, http.StatusBadRequest, "invalid room name (must be between 1 and 120 characters)")
		return
	}

	if req.Name == "" {
		log.Debug().Msg("no room name provided, creating default")
		req.Name = "Debateshare Room"
	}

	// Disabled for now; emails coming soon

	//email, ok := normalizeEmail(req.Email)
	//if !ok {
	//	fail(c, http.StatusBadRequest, "invalid email address")
	//	return
	//}

	var room *models.Room
	for range 5 {
		code, err := models.GenerateCode()
		if err != nil {
			fail(c, http.StatusInternalServerError, "failed to generate room code")
			return
		}
		candidate := &models.Room{
			Code:      code,
			Name:      req.Name,
			ExpiresAt: time.Now().Add(s.cfg.App.RoomTTL),
		}
		err = s.db.Create(candidate).Error
		if err == nil {
			room = candidate
			break
		}
		if !errors.Is(err, gorm.ErrDuplicatedKey) && !isUniqueViolation(err) {
			log.Err(err).Msg("failed to create room")
			fail(c, http.StatusInternalServerError, "failed to create room")
			return
		}
	}
	if room == nil {
		fail(c, http.StatusInternalServerError, "could not allocate a room code")
		return
	}

	// create s3 bucket


	//if email != "" {
	//	s.subscribe(room.ID, email)
	//}
	log.Debug().Str("code", room.Code).Str("name", req.Name).Msg("created room")
	c.JSON(http.StatusCreated, room)
}

func (s *Server) joinRoom(c *gin.Context) {
	room, ok := s.liveRoom(c)
	if !ok {
		return
	}

	var req roomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid json body")
		return
	}
	//email, okEmail := normalizeEmail(req.Email)
	//if !okEmail {
	//	fail(c, http.StatusBadRequest, "invalid email address")
	//	return
	//}
	//if email != "" {
	//	s.subscribe(room.ID, email)
	//}

	log.Debug().Str("code", room.Code).Str("name", req.Name).Msg("joining room")
	c.JSON(http.StatusOK, room)
}

func (s *Server) getRoom(c *gin.Context) {
	room, ok := s.liveRoom(c)
	if !ok {
		return
	}
	files, err := s.roomFiles(room.ID)
	if err != nil {
		fail(c, http.StatusInternalServerError, "failed to get room files")
		return
	}

	log.Debug().Int("files", len(files)).Str("code", room.Code).Msg("got room files")
	c.JSON(http.StatusOK, gin.H{
		"code":        room.Code,
		"name":        room.Name,
		"expiratesAt": room.ExpiresAt,
		"files":       files,
	})

}

func (s *Server) liveRoom(c *gin.Context) (*models.Room, bool) {
	code := strings.ToLower(strings.TrimSpace(c.Param("code")))
	room, err := models.LiveRoom(s.db, code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fail(c, http.StatusNotFound, "room not found or expired")
		} else {
			log.Err(err).
				Str("code", code).
				Msg("failed to load room")
			fail(c, http.StatusInternalServerError, "failed to load room")
		}
		return nil, false
	}
	return room, true
}

func (s *Server) roomFiles(roomID uint) ([]models.File, error) {
	var files []models.File
	err := s.db.Where("room_id = ?", roomID).Order("created_at ASC").Find(&files).Error
	return files, err
}

func (s *Server) subscribe(roomID uint, email string) {
	err := s.db.Clauses(clause.OnConflict{DoNothing: true}).
		Create(&models.Subscriber{RoomID: roomID, Email: email}).Error
	if err != nil {
		log.Err(err).Msgf("failed to subscribe %s to room %d", email, roomID)
	}
}

func normalizeEmail(email string) (string, bool) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return "", false
	}
	if len(email) > 254 || !emailRe.MatchString(email) {
		return "", false
	}
	return email, true
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "duplicate key")
}

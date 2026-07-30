package api

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/imide/debateshare/internal/models"
	"github.com/imide/debateshare/internal/sse"
	"github.com/imide/debateshare/internal/storage"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

func (s *Server) listFiles(c *gin.Context) {
	room, ok := s.liveRoom(c)
	if !ok {
		return
	}

	files, err := s.roomFiles(room.ID)
	if err != nil {
		fail(c, http.StatusInternalServerError, "failed to get room files")
		return
	}

	c.JSON(http.StatusOK, files)
}

func (s *Server) uploadFile(c *gin.Context) {
	room, ok := s.liveRoom(c)
	if !ok {
		return
	}
	header, name, ok := s.readUpload(c)
	if !ok {
		return
	}

	file := &models.File{
		RoomID:      room.ID,
		UUID:        uuid.NewString(),
		Name:        name,
		Size:        header.Size,
		ContentType: contentType(header),
	}
	if !s.storeUpload(c, room.Code, file, header) {
		return
	}
	if err := s.db.Create(file).Error; err != nil {
		log.Err(err).Msg("failed to save file record")
		fail(c, http.StatusInternalServerError, "could not save file")
		return
	}

	//	s.notifyChange(room, "uploaded", file.Name)
	s.hub.Broadcast(room.Code, sse.Event{Name: "file_added", Data: fmt.Sprintf(`{"fileId":%d}`, file.ID)})
	c.JSON(http.StatusCreated, file)
}

func (s *Server) replaceFile(c *gin.Context) {
	room, ok := s.liveRoom(c)
	if !ok {
		return
	}
	file, ok := s.roomFile(c, room.ID)
	if !ok {
		return
	}
	header, name, ok := s.readUpload(c)
	if !ok {
		return
	}

	oldKey := storage.FileKey(room.Code, file.UUID, file.Name)
	file.UUID = uuid.NewString()
	file.Name = name
	file.Size = header.Size
	file.ContentType = contentType(header)
	if !s.storeUpload(c, room.Code, file, header) {
		return
	}
	if err := s.db.Save(file).Error; err != nil {
		log.Err(err).Msg("failed to save file record")
		fail(c, http.StatusInternalServerError, "could not save file")
		return
	}
	if err := s.store.Delete(c.Request.Context(), oldKey); err != nil {
		log.Err(err).Msgf("delete old object %s", oldKey)
	}

	//	s.notifyChange(room, "updated", file.Name)
	s.hub.Broadcast(room.Code, sse.Event{Name: "file_updated", Data: fmt.Sprintf(`{"fileId":%d}`, file.ID)})
	c.JSON(http.StatusOK, file)
}

func (s *Server) deleteFile(c *gin.Context) {
	room, ok := s.liveRoom(c)
	if !ok {
		return
	}
	file, ok := s.roomFile(c, room.ID)
	if !ok {
		return
	}
	if err := s.db.Delete(file).Error; err != nil {
		log.Err(err).Msg("failed to delete file record")
		fail(c, http.StatusInternalServerError, "could not delete file")
		return
	}
	key := storage.FileKey(room.Code, file.UUID, file.Name)
	if err := s.store.Delete(c.Request.Context(), key); err != nil {
		log.Err(err).Msgf("failed to delete %s", key)
	}

	//	s.notifyChange(room, "deleted", file.Name)
	s.hub.Broadcast(room.Code, sse.Event{Name: "file_deleted", Data: fmt.Sprintf(`{"fileId":%d}`, file.ID)})
	c.Status(http.StatusNoContent)
}

func (s *Server) downloadFile(c *gin.Context) {
	room, ok := s.liveRoom(c)
	if !ok {
		return
	}
	file, ok := s.roomFile(c, room.ID)
	if !ok {
		return
	}
	key := storage.FileKey(room.Code, file.UUID, file.Name)
	url, err := s.store.PresignGet(c.Request.Context(), key, file.Name, 24*time.Hour)
	if err != nil {
		log.Err(err).Msgf(
			"failed to create download link for %s/%s", room.Code, file.Name,
		)
		log.Err(err).Str("fileName", file.Name).Str("roomCode", room.Code).Msg("failed to create download link")
		fail(c, http.StatusInternalServerError, "could not create download link")
		return
	}
	c.Redirect(http.StatusFound, url)
}

func (s *Server) readUpload(c *gin.Context) (*multipart.FileHeader, string, bool) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, s.cfg.App.MaxFileSize)
	header, err := c.FormFile("file")
	if err != nil {
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			fail(c, http.StatusRequestEntityTooLarge,
				fmt.Sprintf("file exceeds the %d MB limit", s.cfg.App.MaxFileSize))
		} else {
			fail(c, http.StatusBadRequest, `multipart field "file" is required`)
		}
		return nil, "", false
	}
	name := sanitizeFilename(header.Filename)
	if name == "" {
		fail(c, http.StatusBadRequest, "invalid filename")
		return nil, "", false
	}
	return header, name, true
}

// storeUpload streams the multipart file to S3 under the file's key
func (s *Server) storeUpload(c *gin.Context, roomCode string, file *models.File, header *multipart.FileHeader) bool {
	src, err := header.Open()
	if err != nil {
		fail(c, http.StatusBadRequest, "could not read uploaded file")
		return false
	}
	defer src.Close()
	key := storage.FileKey(roomCode, file.UUID, file.Name)
	if err := s.store.Upload(c.Request.Context(), key, file.ContentType, header.Size, src); err != nil {
		log.Err(err).Str("key", key).Msgf("failed to store file")
		fail(c, http.StatusInternalServerError, "could not store file")
		return false
	}
	return true
}

func (s *Server) roomFile(c *gin.Context, roomID uint) (*models.File, bool) {
	var file models.File
	err := s.db.Where("room_id = ? AND id = ?", roomID, c.Param("id")).First(&file).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fail(c, http.StatusNotFound, "file not found")
		} else {
			fail(c, http.StatusInternalServerError, "could not load file")
		}
		return nil, false
	}
	return &file, true
}

func (s *Server) getRoomZip(c *gin.Context) {
	room, ok := s.liveRoom(c)
	if !ok {
		return
	}
	files, err := s.roomFiles(room.ID)
	if err != nil {
		fail(c, http.StatusInternalServerError, "could not get files for room")
		return
	}

	// tar i know the archive is a .zip, but tar is easier to type. sue me
	tar := storage.ArchiveRoomFiles(c.Request.Context(), s.store, room.Code, files)
	defer tar.Close()

	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.zip"`, room.Code))
	c.Header("Cache-Control", "no-store")
	if _, err := io.Copy(c.Writer, tar); err != nil {
		log.Err(err).
			Str("room", room.Code).
			Msg("failed to stream zip file")
	}
}

// notifyChange emails all subscribers of the room about a file change i
//func (s *Server) notifyChange(room *models.Room, action, fileName string) {
//	var subs []models.Subscriber
//	if err := s.db.Where("room_id = ?", room.ID).Find(&subs).Error; err != nil {
//		log.Err(err).Msgf(
//			"failed to notify subscribers of room %s about %s %s", room.Code, action, fileName,
//		)
//		return
//	}
//	if len(subs) == 0 {
//		return
//	}
//	subject, body := mailer.ChangeEmail(s.Cfg.AppBaseURL, room.Name, room.Code, action, fileName)
//	go func() {
//		for _, sub := range subs {
//			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//			if err := s.mailer.Send(ctx, sub.Email, subject, body, nil); err != nil {
//				log.Err(err).Msgf("failed to send email to %s", sub.Email)
//			}
//			cancel()
//			time.Sleep(600 * time.Millisecond)
//		}
//	}()
//}

func sanitizeFilename(name string) string {
	name = filepath.Base(strings.TrimSpace(name))
	name = strings.Map(func(r rune) rune {
		if r < 32 || r == '/' || r == '\\' {
			return -1
		}
		return r
	}, name)
	if name == "." || name == ".." {
		return ""
	}
	if len(name) > 255 {
		name = name[len(name)-255:]
	}
	return name
}

func contentType(header *multipart.FileHeader) string {
	ct := header.Header.Get("Content-Type")
	if ct == "" {
		ct = "application/octet-stream"
	}
	if len(ct) > 100 {
		ct = ct[:100]
	}
	return ct
}

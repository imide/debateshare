package worker

import (
	"context"
	"time"

	"github.com/imide/debateshare/internal/config"
	"github.com/imide/debateshare/internal/models"
	"github.com/imide/debateshare/internal/sse"
	"github.com/imide/debateshare/internal/storage"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// worker that scans database and store for expired rooms, zips them, and then soft deletes. to prevent data loss, and since it's just documents, the zip files will not be deleted and still served
// even after the room expired at the same URL.
// todo: if the number of unique roomids gets exhausted, just change to alphanumerical system like everyone else... or use squid

const (
	presignTTL      = 7*24*time.Hour - time.Hour // under the sigv4 cap
	attachLimit     = 10 << 20
	sweepInterval   = time.Minute
	perRoomDeadline = 10 * time.Minute
)

type Expirer struct {
	db    *gorm.DB
	store *storage.Store
	hub   *sse.Hub
	cfg   *config.Config
}

func NewExpirer(db *gorm.DB, store *storage.Store, cfg *config.Config, hub *sse.Hub) *Expirer {
	return &Expirer{
		db:    db,
		store: store,
		cfg:   cfg,
		hub:   hub,
	}
}

// sweep for expired rooms every minute and once at startup
func (e *Expirer) Run(ctx context.Context) {
	e.sweep(ctx)
	ticker := time.NewTicker(sweepInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.sweep(ctx)
		}
	}
}

func (e *Expirer) sweep(ctx context.Context) {
	var rooms []models.Room
	if err := e.db.Where("expires_at < ?", time.Now()).Find(&rooms).Error; err != nil {
		log.Err(err).Msg("failed to sweep expired rooms")
		return
	}

	for _, room := range rooms {
		roomCtx, cancel := context.WithTimeout(ctx, perRoomDeadline)
		if err := e.expireRoom(roomCtx, room); err != nil {
			log.Err(err).Msgf("failed to expire room %s (will retry next sweep)", room.Code)
		}
		cancel()
	}
}

func (e *Expirer) expireRoom(ctx context.Context, room models.Room) error {
	log.Info().Msgf("expiring room %s (%q)", room.Code, room.Name)

	var files []models.File
	if err := e.db.Where("room_id = ?", room.ID).Find(&files).Error; err != nil {
		return err
	}
	var subs []models.Subscriber
	if err := e.db.Where("room_id = ?", room.ID).Find(&subs).Error; err != nil {
		return err
	}

	// build and store archive first.
	// failure here aborts and retries next sweep before anything is deleted
	//zipURL := ""
	//var zipSize int64
	if len(files) > 0 {
		zipKey := storage.ArchiveKey(room.Code)
		body := storage.ArchiveRoomFiles(ctx, e.store, room.Code, files)
		err := e.store.UploadStream(ctx, zipKey, "application/zip", body)
		body.Close()
		if err != nil {
			return err
		}
		if _, err = e.store.Size(ctx, zipKey); err != nil {
			return err
		}
		if _, err = e.store.PresignGet(ctx, zipKey, room.Code+".zip", presignTTL); err != nil {
			return err
		}
	}

	// email
	// todo: email

	e.hub.CloseRoom(room.Code, sse.Event{Name: "room_expired", Data: "{}"})

	// soft-delete before S3 cleanup
	if err := e.db.Delete(&room).Error; err != nil {
		return err
	}
	if err := e.db.Where("room_id = ?", room.ID).Delete(&models.File{}).Error; err != nil {
		return err
	}
	//if err := e.store.DeletePrefix(ctx, "rooms/"+room.Code+"/"); err != nil {
	//	log.Err(err).Msgf("cleanup objects for room %s failed", room.Code)
	//}
	log.Info().Msgf("room %s (%q) expired: %d files archived, %d subscribers notified", room.Code, room.Name, len(files), len(subs))

	return nil
}

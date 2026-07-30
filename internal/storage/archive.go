package storage

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/imide/debateshare/internal/models"
	"github.com/rs/zerolog/log"
)

func ArchiveRoomFiles(ctx context.Context, s *Store, roomCode string, files []models.File) io.ReadCloser {
	pr, pw := io.Pipe()
	go func() {
		zw := zip.NewWriter(pw)
		used := map[string]int{}
		for _, f := range files {
			body, err := s.Get(ctx, FileKey(roomCode, f.UUID, f.Name))
			if err != nil {
				log.Err(err).Msgf("archive %s: skipping file %q", roomCode, f.Name)
				continue
			}
			w, err := zw.Create(uniqueName(used, f.Name))
			if err != nil {
				body.Close()
				pw.CloseWithError(err)
				return
			}
			if _, err := io.Copy(w, body); err != nil {
				body.Close()
				pw.CloseWithError(err)
				return
			}
			body.Close()
		}
		pw.CloseWithError(zw.Close())
	}()
	return pr
}

func uniqueName(used map[string]int, name string) string {
	used[name]++
	if used[name] == 1 {
		return name
	}
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	return fmt.Sprintf("%s (%d)%s", base, used[name], ext)
}

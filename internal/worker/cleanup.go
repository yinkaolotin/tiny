package worker

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"github.com/yinkaolotin/tiny/internal/storage"
)

func StartCleanup(ctx context.Context, store storage.Store, log zerolog.Logger) {
	ticker := time.NewTicker(30 * time.Second)

	go func() {
		for {
			select {
			case <-ticker.C:
				count := store.CleanupExpired()
				log.Info().Int("deleted_items", count).Msg("cleanup run")
			case <-ctx.Done():
				log.Info().Msg("cleanup worker stopped")
				ticker.Stop()
				return
			}
		}
	}()
}

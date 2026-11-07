package textprocessing

import (
	"context"
	"customcomplete/internal/i18n"
	"customcomplete/internal/progress"
	"customcomplete/internal/store"
	"database/sql"
	"errors"
	"log/slog"
	"sync"
	"time"
)

var ErrJobAlreadyClaimed = errors.New("already claimed")

func Worker(ctx context.Context, workerID int, readWrite store.ReadWrite, read store.Read, broker *progress.Broker, jobsChan <-chan int64, wg *sync.WaitGroup) {
	defer wg.Done()

	workerSlog := slog.With(slog.Int("worker_id", workerID))
	workerSlog.InfoContext(ctx, "started")

	for {

		select {
		case <-ctx.Done():
			workerSlog.InfoContext(ctx, "stopped", slog.String("reason", "context canceled"))
			return
		case sourceID, ok := <-jobsChan:
			if !ok {
				workerSlog.InfoContext(ctx, "stopped", slog.String("reason", "jobs channel closed"))
				return
			}

			if err := claim(ctx, readWrite, sourceID); err != nil {
				workerSlog.DebugContext(ctx, "claim job",
					slog.Int64("source_id", sourceID),
					slog.Any("err", err),
				)
				continue
			}

			run(ctx, workerSlog, sourceID, readWrite, read, broker)
		case <-time.After(5 * time.Second):

			sourceID, err := claimNext(ctx, readWrite)

			if err != nil {
				if !errors.Is(err, sql.ErrNoRows) && ctx.Err() == nil {
					workerSlog.ErrorContext(ctx, "claim next job", slog.Any("err", err))
				}
				continue
			}

			run(ctx, workerSlog, sourceID, readWrite, read, broker)
		}
	}
}

func claim(ctx context.Context, db store.ReadWrite, sourceID int64) error {
	res, err := db.ExecContext(ctx, `
                UPDATE source 
                SET status = 'processing' 
                WHERE id = ? AND status = 'pending';
            `, sourceID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrJobAlreadyClaimed
	}

	return nil
}

func claimNext(ctx context.Context, db store.ReadWrite) (int64, error) {
	var sourceID int64
	err := db.QueryRowContext(ctx, `
                UPDATE source 
                SET status = 'processing' 
                WHERE id = (
                    SELECT
					id
					FROM source
					WHERE status = 'pending'
					ORDER BY id ASC
					LIMIT 1
                )
                RETURNING id;
            `).Scan(&sourceID)
	return sourceID, err
}

func run(ctx context.Context, wlog *slog.Logger, sourceID int64, readWrite store.ReadWrite, read store.Read, broker *progress.Broker) {
	jobSlog := wlog.With(slog.Int64("source_id", sourceID))
	cleanup := context.WithoutCancel(ctx)

	defer func() {
		if r := recover(); r != nil {
			jobSlog.ErrorContext(cleanup, "process source", slog.Any("panic", r))
			_, err := readWrite.ExecContext(cleanup, `UPDATE source SET status ='failed' WHERE id = ?`, sourceID)
			if err != nil {
				jobSlog.ErrorContext(cleanup, "set status failed after panic", slog.Any("err", err))
			}
			broker.Update(sourceID, 0, i18n.TextStatusFailed)
			return
		}
	}()

	jobSlog.InfoContext(ctx, "started")
	start := time.Now()
	wordCount, processingErr := process(ctx, sourceID, readWrite, read, broker)
	duration := time.Since(start)

	if ctx.Err() != nil {
		jobSlog.InfoContext(cleanup, "interrupted by shutdown", slog.Duration("duration", duration))
		return
	}

	if processingErr != nil {
		jobSlog.ErrorContext(cleanup, "failed",
			slog.Duration("duration", duration),
			slog.Any("err", processingErr),
		)
		if _, err := readWrite.ExecContext(cleanup, `UPDATE source SET status = 'failed' WHERE id = ?`, sourceID); err != nil {
			jobSlog.ErrorContext(cleanup, "update status failed", slog.Any("err", err))
		}
		broker.Update(sourceID, 0, i18n.TextStatusFailed)
		return
	}

	if _, err := readWrite.ExecContext(cleanup, `UPDATE source SET status = 'completed', wordcount = ? WHERE id = ?`, wordCount, sourceID); err != nil {
		jobSlog.ErrorContext(cleanup, "update status completed", slog.Any("err", err))
	}
	broker.Update(sourceID, 100, i18n.TextStatusComplete)
	jobSlog.InfoContext(cleanup, "completed",
		slog.Int("words", wordCount),
		slog.Duration("duration", duration),
	)
}

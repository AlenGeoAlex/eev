package jobs

import (
	"backend-go/internal/services"
	"context"
	"log"
	"time"

	"github.com/robfig/cron/v3"
)

type CleanUpRevokedJob struct {
	cron.Job
	shareableService *services.ShareableService
}

func NewCleanUpRevokedJob(shareableService *services.ShareableService) *CleanUpRevokedJob {
	return &CleanUpRevokedJob{
		shareableService: shareableService,
	}
}

func (c CleanUpRevokedJob) Run() {
	ctx := context.Background()
	log.Println("Cleaning up revoked shareables started at ", time.Now())
	shares, err := c.shareableService.GetExpiredShares(ctx)
	if err != nil {
		log.Println("Failed to clean up revoked shareables: ", err)
		return
	}

	if len(shares) == 0 {
		log.Println("No expired shares found")
		return
	}

	for _, share := range shares {
		log.Printf("Revoking share: %s", share.Name)
		c.shareableService.PublishDeleteShareEvent(services.DeleteShareableEvent{
			ID:         share.ID,
			HardDelete: true,
			Reason:     "AUTOMATED-JOB-[EXPIRY]",
		})
	}

	log.Println("Cleaning up revoked shareables completed at ", time.Now())
}

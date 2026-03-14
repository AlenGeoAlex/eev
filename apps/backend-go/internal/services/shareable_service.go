package services

import (
	"backend-go/internal"
	sqliteeev "backend-go/internal/db/sqlite/generated"
	s3 "backend-go/internal/manager"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"path"
	"time"

	"github.com/oklog/ulid/v2"
)

type ShareableType string

const (
	ShareableTypeText ShareableType = "text"
	ShareableTypeURL  ShareableType = "url"
	ShareableTypeFile ShareableType = "file"
)

type ShareableCode struct {
	ID            string
	Name          string
	UserID        string
	UserEmail     string
	SourceIp      string
	ExpiryAt      time.Time
	ActiveFrom    time.Time
	CreatedAt     time.Time
	ShareableType ShareableType
	ShareableData string
	Options       map[ShareableOption]string
	RevokedAt     *time.Time
}

type ShareableSignedFile struct {
	ID          string
	ShareID     string
	FileName    string
	ContentType string
	ContentSize int64
	SignedURL   string
}

type ShareableOption string

const (
	ShareableOptionExpiryAt                ShareableOption = "expiry_at"
	ShareableOptionOnlyOnce                ShareableOption = "only_once"
	ShareableOptionEncrypt                 ShareableOption = "encrypt"
	ShareableOptionTargetEmails            ShareableOption = "target_emails"
	ShareableOptionEmailNotificationOnOpen ShareableOption = "email_notification_on_open"
	ShareableOptionActiveFrom              ShareableOption = "active_from"
	ShareableOptionAwaitingEncryptedData   ShareableOption = "awaiting_encrypted_data"
)

type ShareableFileDeletionEvent struct {
	ID     string
	S3Keys []string
}

type DeleteShareableEvent struct {
	ID     string
	Reason string
}

type UpdateTargetHistoryEvent struct {
	UserID         string
	EmailAddresses []string
}

type ShareableService struct {
	q         *sqliteeev.Queries
	s3Manager *s3.S3Manager

	// Channels
	FileDeletionEvent        chan ShareableFileDeletionEvent
	DeleteShareableEvent     chan DeleteShareableEvent
	UpdateTargetHistoryEvent chan UpdateTargetHistoryEvent
}

func NewShareableService(q *sqliteeev.Queries, s3Manager *s3.S3Manager) *ShareableService {
	return &ShareableService{
		q:         q,
		s3Manager: s3Manager,

		FileDeletionEvent:        make(chan ShareableFileDeletionEvent, 100),
		DeleteShareableEvent:     make(chan DeleteShareableEvent, 100),
		UpdateTargetHistoryEvent: make(chan UpdateTargetHistoryEvent, 100),
	}
}

func (s ShareableService) InitWorkers(ctx context.Context) {
	s.runFileDeletion(ctx)
	s.runDeleteShareEvent(ctx)
	s.runUpdateTargetHistoryEvent(ctx)
}

func (s ShareableService) PublishFileDeletionEvent(e ShareableFileDeletionEvent) {
	s.FileDeletionEvent <- e
}

func (s ShareableService) PublishDeleteShareEvent(e DeleteShareableEvent) {
	s.DeleteShareableEvent <- e
}

func (s ShareableService) publishUpdateTargetHistoryEvent(e UpdateTargetHistoryEvent) {
	s.UpdateTargetHistoryEvent <- e
}

func (s ShareableService) runUpdateTargetHistoryEvent(ctx context.Context) {
	go func() {
		for {
			select {
			case event := <-s.UpdateTargetHistoryEvent:
				{
					for _, address := range event.EmailAddresses {
						err := s.q.UpsertTargetsForUser(ctx, sqliteeev.UpsertTargetsForUserParams{
							UserID:      event.UserID,
							TargetEmail: address,
						})
						if err != nil {
							log.Printf("Failed to upsert targets for user %s: %v", event.UserID, err)
							return
						}
					}
				}
			case <-ctx.Done():
				log.Printf("Stopping the runUpdateTargetHistoryEvent due to context cancellation. [%s]", ctx.Err())
			}
		}
	}()
}

func (s ShareableService) runDeleteShareEvent(ctx context.Context) {
	go func() {
		for {
			select {
			case event := <-s.DeleteShareableEvent:
				{
					log.Printf("DeleteShareableEvent received for %s with reason %s", event.ID, event.Reason)
					shareableWithOptions, err := s.q.GetShareable(ctx, event.ID)
					if err != nil {
						log.Printf("Error getting shareable for %s. Aborting deletion", event.ID)
						return
					}

					if shareableWithOptions == nil || errors.Is(err, sql.ErrNoRows) || len(shareableWithOptions) == 0 {
						log.Printf("Shareable for %s was nil", event.ID)
						return
					}

					shareable := shareableWithOptions[0] //(Ignore the options with this one)
					if shareable.ShareableType == string(ShareableTypeFile) {
						shareFiles, err := s.q.GetShareableFilesOfShare(ctx, shareable.ID)
						if err == nil && len(shareFiles) > 0 {
							s3Keys := make([]string, len(shareFiles))
							for _, file := range shareFiles {
								s3Keys = append(s3Keys, file.S3Key)
							}

							s.PublishFileDeletionEvent(ShareableFileDeletionEvent{
								ID:     shareable.ID,
								S3Keys: s3Keys,
							})
						} else {
							log.Printf("Error getting shareable for %s. Aborting deletion", event.ID)
						}
					}

					err = s.q.SetShareableAsRevoked(ctx, sqliteeev.SetShareableAsRevokedParams{
						ID:        shareable.ID,
						RevokedAt: sql.NullTime{Time: time.Now(), Valid: true},
					})
					if err != nil {
						log.Printf("Failed to set shareable as revoked for %s: %v", shareable.ID, err)
						return
					}

					log.Printf("Shareable %s revoked", shareable.ID)
				}
			case <-ctx.Done():
				log.Printf("Stopping the runDeleteShareEvent due to context cancellation. [%s]", ctx.Err())
			}
		}
	}()
}

func (s ShareableService) runFileDeletion(ctx context.Context) {
	go func() {
		for {
			select {
			case event := <-s.FileDeletionEvent:
				{
					log.Printf("Received file deletion event: %s with %d urls", event.ID, len(event.S3Keys))

				}
			case <-ctx.Done():
				log.Println("EventBus shutting down for FileDeletionEvent")
			}
		}
	}()
}

func (s ShareableService) CreateShareable(
	ctx context.Context,
	name string,
	userID string,
	sourceIP string,
	shareableType string,
	shareableData *string,
	expiryAt time.Time,
	activeFrom time.Time,
	options map[ShareableOption]string,
) (shareable *ShareableCode, error error) {

	id := generateID()
	data := ""
	if shareableData != nil {
		data = *shareableData
	}
	err := s.q.InsertShareable(ctx, sqliteeev.InsertShareableParams{
		ID:            id,
		Name:          name,
		UserID:        userID,
		SourceIp:      sql.NullString{String: sourceIP, Valid: sourceIP != ""},
		ExpiryAt:      expiryAt,
		ActiveFrom:    activeFrom,
		ShareableType: shareableType,
		ShareableData: data,
	})
	if err != nil {
		return nil, errors.New("failed to create shareable")
	}

	for option, val := range options {
		err := s.upsertShareableOption(option, val, id)
		if err != nil {
			s.PublishDeleteShareEvent(DeleteShareableEvent{
				ID: id,
			})
			return nil, errors.New("failed to create shareable option - " + string(option))
		}
	}

	targetEmailsJson := options[ShareableOptionTargetEmails]
	s.updateTargetEmails(targetEmailsJson, userID)

	shareable = &ShareableCode{
		ID:            id,
		Name:          name,
		UserID:        userID,
		UserEmail:     "",
		SourceIp:      sourceIP,
		ExpiryAt:      expiryAt,
		CreatedAt:     time.Now(),
		ShareableType: ShareableType(shareableType),
		ShareableData: data,
		Options:       options,
	}

	return shareable, nil
}

func (s ShareableService) updateTargetEmails(targetEmailsJson string, userId string) {
	var emails []string

	err := json.Unmarshal([]byte(targetEmailsJson), &emails)
	if err != nil {
		log.Printf("Error unmarshalling emails json: %v", err)
		return
	}

	s.publishUpdateTargetHistoryEvent(UpdateTargetHistoryEvent{
		UserID:         userId,
		EmailAddresses: emails,
	})
	log.Printf("Updated target emails for %s with %d emails", userId, len(emails))
}

func (s ShareableService) upsertShareableOption(
	option ShareableOption,
	val string,
	shareId string,
) (err error) {
	err = s.q.UpsertShareableOption(context.Background(), sqliteeev.UpsertShareableOptionParams{
		ShareID:   shareId,
		OptionKey: string(option),
		Value:     val,
	})
	if err != nil {
		log.Printf("Failed to upsert shareable option: %s", err)
		return err
	}

	return nil
}

func (s ShareableService) CreateShareableFile(
	ctx context.Context,
	shareId string,
	userId string,
	fileName string,
	fileSize int64,
	contentType string,
) (id string, signedURL string, error error) {
	id = ulid.Make().String()
	ext := path.Ext(fileName)
	var s3Key = "shareable/" + userId + "/" + shareId + "/" + id + ext

	err := s.q.InsertShareableFile(ctx, sqliteeev.InsertShareableFileParams{
		ID:          id,
		ShareID:     shareId,
		FileName:    fileName,
		ContentType: contentType,
		S3Key:       s3Key,
		ContentSize: float64(fileSize),
	})
	if err != nil {
		return id, "", err
	}

	signedURL, err = s.s3Manager.PresignPutObject(ctx, s3Key, contentType, fileSize)
	if err != nil {
		return id, "", err
	}

	log.Printf("Generated signed URL for file %s: %s", fileName, signedURL)
	return id, signedURL, nil
}

func generateID() string {
	return internal.MicroTimeID()
}

func (s ShareableService) GetSignedFilesForShareable(id string) ([]ShareableSignedFile, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}

	files, err := s.q.GetShareableFiles(context.Background(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return make([]ShareableSignedFile, 0), nil
		}
		return nil, err
	}

	filesWithSignedURLs := make([]ShareableSignedFile, len(files))
	for i, file := range files {
		signedURL, err := s.s3Manager.PresignGetObject(context.Background(), file.S3Key)
		if err != nil {
			return nil, errors.Join(errors.New("Failed to generate signed url for "+file.S3Key), err)
		}

		filesWithSignedURLs[i] = ShareableSignedFile{
			ID:          file.ID,
			ShareID:     file.ShareID,
			FileName:    file.FileName,
			ContentType: file.ContentType,
			ContentSize: int64(file.ContentSize),
			SignedURL:   signedURL,
		}
	}

	return filesWithSignedURLs, nil
}

func (s ShareableService) GetShareableInfoFromCode(
	code string,
) (*ShareableCode, error) {

	if code == "" {
		return nil, errors.New("code cannot be empty")
	}

	shareable, err := s.q.GetShareable(context.Background(), code)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	shareableCode, err := toShareableCode(&shareable)
	if err == nil && shareableCode == nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if shareableCode.RevokedAt != nil {
		return nil, errors.New("shareable code is revoked")
	}

	if shareableCode.ActiveFrom.After(time.Now()) {
		return shareableCode, errors.New("shareable code is not active yet")
	}

	if shareableCode.ExpiryAt.Before(time.Now()) {
		return nil, errors.New("shareable code has expired")
	}

	return shareableCode, nil
}

func toShareableCode(shareable *[]sqliteeev.GetShareableRow) (*ShareableCode, error) {
	if len(*shareable) == 0 {
		return nil, nil
	}

	var revokedAt *time.Time = nil
	if (*shareable)[0].RevokedAt.Valid {
		revokedAt = &(*shareable)[0].RevokedAt.Time
	}
	shareableCode := &ShareableCode{
		ID:            (*shareable)[0].ID,
		Name:          (*shareable)[0].Name,
		UserID:        (*shareable)[0].UserID,
		UserEmail:     (*shareable)[0].Email,
		ExpiryAt:      (*shareable)[0].ExpiryAt,
		ShareableType: ShareableType((*shareable)[0].ShareableType),
		ShareableData: (*shareable)[0].ShareableData,
		ActiveFrom:    (*shareable)[0].ActiveFrom,
		Options:       map[ShareableOption]string{},
		RevokedAt:     revokedAt,
	}

	for _, row := range *shareable {
		shareableCode.Options[ShareableOption(row.OptionKey)] = row.Value
	}

	return shareableCode, nil
}

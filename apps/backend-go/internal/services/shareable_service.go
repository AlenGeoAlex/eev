package services

import (
	"backend-go/internal"
	sqliteeev "backend-go/internal/db/sqlite/generated"
	s3 "backend-go/internal/manager"
	"context"
	"database/sql"
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
	ExpiryAt      *time.Time
	CreatedAt     time.Time
	ShareableType string
	ShareableData string
	Options       map[ShareableOption]string
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

type ShareableService struct {
	q         *sqliteeev.Queries
	s3Manager *s3.S3Manager
}

func NewShareableService(q *sqliteeev.Queries, s3Manager *s3.S3Manager) *ShareableService {
	return &ShareableService{q: q, s3Manager: s3Manager}
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
			// TODO PUSH TO DELETE IN CHANNELS
			return nil, errors.New("failed to create shareable option - " + string(option))
		}
	}

	shareable = &ShareableCode{
		ID:            id,
		Name:          name,
		UserID:        userID,
		UserEmail:     "",
		SourceIp:      sourceIP,
		ExpiryAt:      &expiryAt,
		CreatedAt:     time.Now(),
		ShareableType: shareableType,
		ShareableData: data,
		Options:       options,
	}

	return shareable, nil
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

func (s ShareableService) GetShareableInfoFromCode(
	code string,
) (*ShareableCode, error) {

	if code == "" {
		return nil, errors.New("code cannot be empty")
	}

	shareable, err := s.q.GetShareable(context.Background(), code)
	if err != nil {
		return nil, err
	}

	shareableCode, err := toShareableCode(&shareable)
	if err == nil && shareableCode == nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return shareableCode, nil
}

func toShareableCode(shareable *[]sqliteeev.GetShareableRow) (*ShareableCode, error) {
	if len(*shareable) == 0 {
		return nil, nil
	}

	shareableCode := &ShareableCode{
		ID:            (*shareable)[0].ID,
		Name:          (*shareable)[0].Name,
		UserID:        (*shareable)[0].UserID,
		UserEmail:     (*shareable)[0].Email,
		ExpiryAt:      nil,
		ShareableType: (*shareable)[0].ShareableType,
		ShareableData: (*shareable)[0].ShareableData,
		Options:       map[ShareableOption]string{},
	}

	shareableCode.ExpiryAt = &(*shareable)[0].ExpiryAt

	for _, row := range *shareable {
		shareableCode.Options[ShareableOption(row.OptionKey)] = row.Value
	}

	return shareableCode, nil
}

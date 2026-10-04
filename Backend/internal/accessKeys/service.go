package accesskeys

import (
	"context"
	"crypto/rand"
	"log"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/google/uuid"
)

type AccessKeyStore interface {
	Create(context.Context, *model.AccessKey) error
	GetByID(context.Context, uuid.UUID) (model.AccessKey, error)
	Redeem(context.Context, CodeOfUser) error
	Revoke(context.Context, RevokeCodeData) error
}

type SendEmail interface {
	SendActivationCode(ctx context.Context, to, code string, expiresAt time.Time) error
}

type CodeOfUser struct {
	user_Id uuid.UUID
	code    string
}
type AccessKeyService struct {
	store AccessKeyStore
	email SendEmail
}

func NewAccessKeyService(store AccessKeyStore, email SendEmail) *AccessKeyService {
	return &AccessKeyService{store, email}
}

func (s *AccessKeyService) IssueForUser(ctx context.Context, municipalID uuid.UUID, email string) (model.AccessKey, error) {
	expire := time.Now().AddDate(1, 0, 0)
	code := rand.Text()

	accessKey := &model.AccessKey{
		Code:      code,
		CreatedBy: municipalID,
		ExpiresAt: &expire,
	}

	if err := s.store.Create(ctx, accessKey); err != nil {
		return model.AccessKey{}, err
	}

	s.sendEmailAsync(email, code, expire)

	return *accessKey, nil

}

func (s *AccessKeyService) Resend(ctx context.Context, accessKeyId uuid.UUID, email string) error {
	accessKey, err := s.store.GetByID(ctx, accessKeyId)

	if err != nil {
		return err
	}

	s.sendEmailAsync(email, accessKey.Code, *accessKey.ExpiresAt)

	return nil
}

func (s *AccessKeyService) RevokePremature(ctx context.Context, municipalID uuid.UUID, code string) error {
	revoke := &RevokeCodeData{
		RevokeBy: municipalID,
		code:     code,
	}

	if err := s.store.Revoke(ctx, *revoke); err != nil {
		return err
	}

	return nil
}

func (s *AccessKeyService) sendEmailAsync(to, code string, expire time.Time) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := s.email.SendActivationCode(ctx, to, code, expire); err != nil {
			log.Printf("error enviando correo de actvación a %s: %v", to, err)
		}
	}()
}

func (s *AccessKeyService) Redeem(ctx context.Context, code string, userID uuid.UUID) error {
	codeUser := &CodeOfUser{
		user_Id: userID,
		code:    code,
	}

	if err := s.store.Redeem(ctx, *codeUser); err != nil {
		return err
	}

	return nil
}

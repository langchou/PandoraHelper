package service

import (
	"PandoraHelper/internal/model"
	"PandoraHelper/internal/repository"
	"context"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type ShareService interface {
	RefreshShareToken(accessToken string, share *model.Share, resetLimit bool) (string, error)
	GetShare(ctx context.Context, id int64) (*model.Share, error)
	Update(ctx context.Context, share *model.Share) error
	Create(ctx context.Context, share *model.Share) error
	SearchShare(ctx context.Context, email string, uniqueName string) ([]*model.Share, error)
	DeleteShare(ctx context.Context, id int64) error
}

func NewShareService(service *Service, shareRepository repository.ShareRepository, viper *viper.Viper, accountService AccountService) ShareService {
	return &shareService{
		Service:         service,
		shareRepository: shareRepository,
		viper:           viper,
		accountService:  accountService,
	}
}

type shareService struct {
	*Service
	shareRepository repository.ShareRepository
	viper           *viper.Viper
	accountService  AccountService
}

func (s *shareService) RefreshShareToken(accessToken string, share *model.Share, resetLimit bool) (string, error) {
	chatDomain := fmt.Sprintf("%s/token/register", s.viper.GetString("pandora.domain.chat"))
	var resp struct {
		TokenKey string `json:"token_key"`
	}
	client := resty.New()
	_, err := client.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetFormData(map[string]string{
			"unique_name":        share.UniqueName,
			"access_token":       accessToken,
			"expires_in":         fmt.Sprintf("%d", share.ExpiresIn),
			"site_limit":         share.SiteLimit,
			"reset_limit":        fmt.Sprintf("%t", resetLimit),
			"show_conversations": fmt.Sprintf("%t", !share.ShowConversations),
			"show_userinfo":      fmt.Sprintf("%t", share.ShowUserinfo),
			"gpt35_limit":        fmt.Sprintf("%d", share.Gpt35Limit),
			"gpt4_limit":         fmt.Sprintf("%d", share.Gpt4Limit),
		}).
		SetResult(&resp).
		Post(chatDomain)
	if err != nil {
		s.logger.Error("RefreshShareToken error", zap.Any("err", err))
		return "", err
	}
	s.logger.Info("RefreshShareToken resp", zap.Any("resp", resp))
	return resp.TokenKey, nil
}

func (s *shareService) Update(ctx context.Context, share *model.Share) error {
	err := s.shareRepository.Update(ctx, share)
	if err != nil {
		return err
	}
	return nil
}

func (s *shareService) Create(ctx context.Context, share *model.Share) error {
	err := s.shareRepository.Create(ctx, share)
	if err != nil {
		return err
	}
	s.RefreshShareToken(share.Account.AccessToken, share, true)
	return nil
}

func (s *shareService) SearchShare(ctx context.Context, email string, uniqueName string) ([]*model.Share, error) {
	return s.shareRepository.SearchShare(ctx, email, uniqueName)
}

func (s *shareService) DeleteShare(ctx context.Context, id int64) error {
	return s.shareRepository.DeleteShare(ctx, id)
}

func (s *shareService) GetShare(ctx context.Context, id int64) (*model.Share, error) {
	return s.shareRepository.GetShare(ctx, id)
}

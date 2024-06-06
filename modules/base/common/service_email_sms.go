package common

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/config"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/log"
	"go.uber.org/zap"
)

type IEmailProvider interface {
	SendEmail(ctx context.Context, email string, code string) error
}

// IEmailService IEmailService
type IEmailService interface {
	// 发送验证码
	SendVerifyCode(ctx context.Context, email string, codeType CodeType) error
	// 验证验证码(销毁缓存)
	Verify(ctx context.Context, email, code string, codeType CodeType) error
}

// EmailService 邮箱服务
type EmailService struct {
	ctx *config.Context
	log.Log
}

// NewEmailService 创建短信服务
func NewEmailService(ctx *config.Context) *EmailService {
	return &EmailService{
		ctx: ctx,
		Log: log.NewTLog("EmailService"),
	}
}

// SendVerifyCode 发送邮箱验证码
func (s *EmailService) SendVerifyCode(ctx context.Context, email string, codeType CodeType) error {
	var emailProvider IEmailProvider
	verifyCode := ""
	rand.Seed(int64(time.Now().Nanosecond()))
	for i := 0; i < 4; i++ {
		verifyCode += fmt.Sprintf("%v", rand.Intn(10))
	}
	s.Info("发送验证码", zap.String("code", verifyCode))
	cacheKey := fmt.Sprintf("%s%d@%s@%s", CacheKeyEmailCode, codeType, email)
	err := s.ctx.GetRedisConn().SetAndExpire(cacheKey, verifyCode, time.Minute*5)
	if err != nil {
		return err
	}
	err = emailProvider.SendEmail(ctx, email, verifyCode)
	return err
}

// Verify 验证验证码
func (s *EmailService) Verify(ctx context.Context, email, code string, codeType CodeType) error {
	span, _ := s.ctx.Tracer().StartSpanFromContext(ctx, "emailService.Verify")
	defer span.Finish()
	cacheKey := fmt.Sprintf("%s%d@%s@%s", CacheKeyEmailCode, codeType, email)
	sysCode, err := s.ctx.GetRedisConn().GetString(cacheKey)
	if err != nil {
		return err
	}
	if sysCode != "" && sysCode == code {
		s.ctx.GetRedisConn().Del(cacheKey)
		return nil
	}
	return errors.New("验证码无效！")
}

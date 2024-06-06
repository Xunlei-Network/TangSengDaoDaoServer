package common

import (
	"context"
	"fmt"
	"github.com/opentracing/opentracing-go/ext"
	"net/smtp"
	"strings"
	"time"

	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/config"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/log"
)

type AliyunEmailProvider struct {
	ctx *config.Context
	log.Log
}

// NewAliyunProvider 创建短信服务
func NewAliyunEmailProvider(ctx *config.Context) IEmailProvider {
	return &AliyunEmailProvider{
		ctx: ctx,
		Log: log.NewTLog("AliyunEmailProvider"),
	}
}

func (a *AliyunEmailProvider) SendEmail(ctx context.Context, email string, code string) error {
	fmt.Println("AliyunEmailProvider......")
	span, _ := a.ctx.Tracer().StartSpanFromContext(ctx, "emailService.SendVerifyCode")
	defer span.Finish()

	user := a.ctx.GetConfig().FromUser         // "notify@mailsboard.com"
	password := a.ctx.GetConfig().FromPassword // "XunLei102478901"
	host := a.ctx.GetConfig().Host             //"smtpdm-ap-southeast-1.aliyun.com:80"
	//收件人地址，收件人地址1…………
	to := []string{email} //[]string{"zjycry9312@163.com"}
	subject := "【重庆石嘉千汉网络科技】邮箱验证码验证"
	date := fmt.Sprintf("%s", time.Now().Format(time.RFC1123Z))
	mailtype := "html"
	//回复地址
	replyToAddress := ""
	body := fmt.Sprintln("<html><body><h3>【重庆石嘉千汉网络科技】你的动态码为：{0}，5分钟内有效，请勿泄漏！</h3></body></html>", code)
	fmt.Println("send email")
	err := SendToMail(user, password, host, subject, date, body, mailtype, replyToAddress, to, nil, nil)
	if err != nil {
		ext.LogError(span, err)
		return err
		//panic("邮箱验证码发送失败！")
	}
	fmt.Println("邮箱验证码发送验证码成功......ok...")
	return nil
}

// 发送邮箱验证码接口
func SendToMail(user, password, host, subject, date, body, mailtype, replyToAddress string, to, cc, bcc []string) error {
	hp := strings.Split(host, ":")
	auth := smtp.PlainAuth("", user, password, hp[0])
	var content_type string
	if mailtype == "html" {
		content_type = "Content-Type: text/" + mailtype + "; charset=UTF-8"
	} else {
		content_type = "Content-Type: text/plain" + "; charset=UTF-8"
	}

	cc_address := strings.Join(cc, ";")
	bcc_address := strings.Join(bcc, ";")
	to_address := strings.Join(to, ";")
	msg := []byte("To: " + to_address + "\r\nFrom: " + user + "\r\nSubject: " + subject + "\r\nDate: " + date + "\r\nReply-To: " + replyToAddress + "\r\nCc: " + cc_address + "\r\nBcc: " + bcc_address + "\r\n" + content_type + "\r\n\r\n" + body)
	send_to := MergeSlice(to, cc)
	send_to = MergeSlice(send_to, bcc)
	err := smtp.SendMail(host, auth, user, send_to, msg)
	return err
}

func MergeSlice(s1 []string, s2 []string) []string {
	slice := make([]string, len(s1)+len(s2))
	copy(slice, s1)
	copy(slice[len(s1):], s2)
	return slice
}

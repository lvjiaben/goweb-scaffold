package common

import (
	"net/http"

	"github.com/lvjiaben/goweb-core/errorsx"
	"github.com/lvjiaben/goweb-core/httpx"
	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
)

type Module struct{}

func (Module) Name() string { return "common" }

func (Module) Register(runtime *bootstrap.Runtime) error {
	runtime.AdminCommonGroup.POST("/captcha", captcha(runtime))
	runtime.AppCommonGroup.POST("/captcha", captcha(runtime))
	return nil
}

func captcha(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	return func(c *httpx.Context) {
		if runtime == nil {
			c.Fail(http.StatusInternalServerError, errorsx.CodeInternal, "验证码服务未初始化", map[string]any{})
			return
		}
		if runtime.CaptchaService == nil {
			runtime.Logger.Error("captcha service is nil")
			c.Fail(http.StatusInternalServerError, errorsx.CodeInternal, "验证码服务未初始化", map[string]any{})
			return
		}
		captchaID, captchaData, err := runtime.CaptchaService.Generate()
		if err != nil {
			runtime.Logger.Error("captcha generate failed", "error", err)
			c.Error(err)
			return
		}
		c.Success(map[string]any{
			"captcha_id":   captchaID,
			"captcha_data": captchaData,
		})
	}
}

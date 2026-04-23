package common

import (
	"github.com/lvjiaben/goweb-core/httpx"
	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
)

type Module struct{}

func (Module) Name() string { return "common" }

func (Module) Register(runtime *bootstrap.Runtime) error {
	runtime.BackendPublicGroup.Group("/common").POST("/captcha", captcha(runtime))
	runtime.AppPublicGroup.Group("/common").POST("/captcha", captcha(runtime))
	return nil
}

func captcha(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	return func(c *httpx.Context) {
		captchaID, captchaData, err := runtime.CaptchaService.Generate()
		if err != nil {
			c.Error(err)
			return
		}
		c.Success(map[string]any{
			"captcha_id":   captchaID,
			"captcha_data": captchaData,
		})
	}
}

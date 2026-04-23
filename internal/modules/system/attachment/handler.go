package attachment

import (
	"github.com/lvjiaben/goweb-core/httpx"
	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
)

func upload(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	service := NewService(runtime)
	return func(c *httpx.Context) {
		file, header, err := c.MultipartFormFile("file")
		if err != nil {
			c.Error(err)
			return
		}
		uploaderID := int64(0)
		if adminUser, _ := bootstrap.CurrentAdminUser(c); adminUser != nil {
			uploaderID = adminUser.ID
		}
		result, err := service.Upload(UploadRequest{
			File:         file,
			Header:       header,
			Parent:       c.Request.FormValue("parent"),
			UploaderID:   uploaderID,
			UploaderKind: "admin",
		})
		if err != nil {
			c.Error(err)
			return
		}
		c.Success(result)
	}
}

func directories(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	service := NewService(runtime)
	return func(c *httpx.Context) {
		result, err := service.Directories()
		if err != nil {
			c.Error(err)
			return
		}
		c.Success(result)
	}
}

func list(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	service := NewService(runtime)
	return func(c *httpx.Context) {
		page, pageSize := bootstrap.Pagination(c)
		result, err := service.List(ListParams{
			Page:     page,
			PageSize: pageSize,
			Keyword:  bootstrap.SearchKeyword(c),
			Parent:   bootstrap.QueryFirst(c, "parent", "directory"),
		})
		if err != nil {
			c.Error(err)
			return
		}
		c.Success(result)
	}
}

func deleteFiles(runtime *bootstrap.Runtime) httpx.HandlerFunc {
	service := NewService(runtime)
	return func(c *httpx.Context) {
		var req bootstrap.IDsPayload
		if err := c.BindJSON(&req); err != nil {
			c.Error(err)
			return
		}
		result, err := service.DeleteFiles(req.Values())
		if err != nil {
			respondServiceError(c, err)
			return
		}
		c.Success(result)
	}
}

func respondServiceError(c *httpx.Context, err error) {
	if isValidationError(err) {
		c.BadRequest(err.Error())
		return
	}
	c.Error(err)
}

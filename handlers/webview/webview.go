package webview

import (
	"github.com/labstack/echo/v4"
)

type WebviewHandler struct{}

func New() *WebviewHandler {
	return &WebviewHandler{}
}

func GetName() string {
	return "webview"
}

func (h *WebviewHandler) RenderBid(c echo.Context) error {
	// return c.HTML(http.StatusOK, "<div><title>Place bid</title><body>Hello world!</body></div>")
	// http.Handle("/", http.FileServer(http.Dir("./static")))
	return nil
}

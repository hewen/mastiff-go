// Package unicontext provides a context interface for HTTP handlers.
package unicontext

import (
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GinContext implements the UniversalContext interface for Gin.
type GinContext struct {
	Ctx *gin.Context
}

// Request returns the HTTP request.
func (c *GinContext) Request() *http.Request {
	return c.Ctx.Request
}

// ResponseWriter returns the HTTP response writer.
func (c *GinContext) ResponseWriter() http.ResponseWriter {
	return c.Ctx.Writer
}

// Next calls the next handler in the chain.
func (c *GinContext) Next() error {
	c.Ctx.Next()
	return nil
}

// Param returns the value of the URL parameter with the given key.
// It returns an empty string if the key does not exist.
func (c *GinContext) Param(key string) string {
	return c.Ctx.Param(key)
}

// Query returns the value of the URL query parameter with the given key.
// It returns an empty string if the key does not exist.
func (c *GinContext) Query(key string) string {
	return c.Ctx.Query(key)
}

// Header returns the value of the HTTP header with the given key.
// It returns an empty string if the key does not exist.
func (c *GinContext) Header(key string) string {
	return c.Ctx.GetHeader(key)
}

// Cookie returns the value of the HTTP cookie with the given key.
// It returns an empty string if the key does not exist.
func (c *GinContext) Cookie(key string) string {
	val, _ := c.Ctx.Cookie(key)
	return val
}

// JSON sends a JSON response with the given status code and data.
func (c *GinContext) JSON(status int, data any) error {
	c.Ctx.JSON(status, data)
	return nil
}

// Text sends a text response with the given status code and text.
func (c *GinContext) Text(status int, text string) error {
	c.Ctx.String(status, text)
	return nil
}

// String sends a string response with the given status code and formatted text.
func (c *GinContext) String(status int, format string, values ...any) error {
	c.Ctx.String(status, format, values...)
	return nil
}

// HTML sends an HTML response with the given status code and HTML template name and data.
func (c *GinContext) HTML(status int, name string, obj any) error {
	c.Ctx.HTML(status, name, obj)
	return nil
}

// Redirect sends a redirect response with the given status code and URL.
func (c *GinContext) Redirect(status int, url string) error {
	c.Ctx.Redirect(status, url)
	return nil
}

// File sends a file response with the given filepath.
func (c *GinContext) File(filepath string) error {
	c.Ctx.File(filepath)
	return nil
}

// Attachment sends an attachment response with the given filepath and filename.
func (c *GinContext) Attachment(filepath, filename string) error {
	c.Ctx.FileAttachment(filepath, filename)
	return nil
}

// BindJSON binds the JSON request body into the given object.
func (c *GinContext) BindJSON(obj any) error {
	return c.Ctx.ShouldBindJSON(obj)
}

// FormValue returns the value of the form field with the given key.
// It returns an empty string if the key does not exist.
func (c *GinContext) FormValue(key string) string {
	return c.Ctx.PostForm(key)
}

// FormFile returns the file header of the form file with the given key.
// It returns an error if the key does not exist.
func (c *GinContext) FormFile(key string) (*multipart.FileHeader, error) {
	return c.Ctx.FormFile(key)
}

// Method returns the HTTP method of the request.
func (c *GinContext) Method() string {
	return c.Ctx.Request.Method
}

// Path returns the path of the request.
func (c *GinContext) Path() string {
	return c.Ctx.Request.URL.Path
}

// FullPath returns the path of the request.
func (c *GinContext) FullPath() string {
	return c.Ctx.FullPath()
}

// ClientIP returns the client IP of the request.
func (c *GinContext) ClientIP() string {
	return c.Ctx.ClientIP()
}

// Set sets the value of the context with the given key.
func (c *GinContext) Set(key string, value any) {
	c.Ctx.Set(key, value)
}

// Get returns the value of the context with the given key.
// It returns nil if the key does not exist.
func (c *GinContext) Get(key string) (any, bool) {
	return c.Ctx.Get(key)
}

// StatusCode returns the status code of the response.
func (c *GinContext) StatusCode() int {
	return c.Ctx.Writer.Status()
}

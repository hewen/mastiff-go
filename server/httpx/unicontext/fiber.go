// Package unicontext provides a context interface for HTTP handlers.
package unicontext

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

// FiberContext implements the UniversalContext interface for Fiber.
type FiberContext struct {
	Ctx *fiber.Ctx
	rw  *fiberResponseWriter
}

// Request returns the HTTP request.
func (c *FiberContext) Request() *http.Request {
	req, _ := http.NewRequestWithContext(
		c.Ctx.Context(),
		c.Ctx.Method(),
		c.Ctx.OriginalURL(),
		bytes.NewReader(c.Ctx.Body()),
	)

	c.Ctx.Request().Header.VisitAll(func(key, value []byte) {
		req.Header.Set(string(key), string(value))
	})

	req.RemoteAddr = c.Ctx.Context().RemoteAddr().String()

	return req
}

// ResponseWriter returns the HTTP response writer.
func (c *FiberContext) ResponseWriter() http.ResponseWriter {
	if c.rw == nil {
		c.rw = &fiberResponseWriter{
			ctx:    c.Ctx,
			header: http.Header{},
		}
	}
	return c.rw
}

// Next calls the next handler in the chain.
func (c *FiberContext) Next() error {
	return c.Ctx.Next()
}

// Param returns the value of the URL parameter with the given key.
// It returns an empty string if the key does not exist.
func (c *FiberContext) Param(key string) string {
	return c.Ctx.Params(key)
}

// Query returns the value of the URL query parameter with the given key.
// It returns an empty string if the key does not exist.
func (c *FiberContext) Query(key string) string {
	return c.Ctx.Query(key)
}

// Header returns the value of the HTTP header with the given key.
// It returns an empty string if the key does not exist.
func (c *FiberContext) Header(key string) string {
	return c.Ctx.Get(key)
}

// Cookie returns the value of the HTTP cookie with the given key.
// It returns an empty string if the key does not exist.
func (c *FiberContext) Cookie(key string) string {
	return c.Ctx.Cookies(key)
}

// Data writes some data into the body stream and updates the HTTP code.
func (c *FiberContext) Data(status int, contentType string, data []byte) error {
	c.Ctx.Context().SetContentType(contentType)
	return c.Ctx.Status(status).Send(data)
}

// JSON sends a JSON response with the given status code and data.
func (c *FiberContext) JSON(status int, data any) error {
	return c.Ctx.Status(status).JSON(data)
}

// Text sends a text response with the given status code and text.
func (c *FiberContext) Text(status int, text string) error {
	return c.Ctx.Status(status).SendString(text)
}

// String sends a string response with the given status code and formatted text.
func (c *FiberContext) String(status int, format string, values ...any) error {
	return c.Ctx.Status(status).SendString(fmt.Sprintf(format, values...))
}

// HTML sends an HTML response with the given status code and HTML template name and data.
func (c *FiberContext) HTML(status int, name string, obj any) error {
	return c.Ctx.Status(status).Render(name, obj)
}

// Redirect sends a redirect response with the given status code and URL.
func (c *FiberContext) Redirect(status int, url string) error {
	return c.Ctx.Redirect(url, status)
}

// File sends a file response with the given filepath.
func (c *FiberContext) File(filepath string) error {
	return c.Ctx.SendFile(filepath)
}

// Attachment sends an attachment response with the given filepath and filename.
func (c *FiberContext) Attachment(filepath, filename string) error {
	c.Ctx.Attachment(filename)
	return c.Ctx.SendFile(filepath)
}

// BindJSON binds the JSON request body into the given object.
func (c *FiberContext) BindJSON(obj any) error {
	return c.Ctx.BodyParser(obj)
}

// FormValue returns the value of the form field with the given key.
// It returns an empty string if the key does not exist.
func (c *FiberContext) FormValue(key string) string {
	return c.Ctx.FormValue(key)
}

// FormFile returns the file header of the form file with the given key.
// It returns an error if the key does not exist.
func (c *FiberContext) FormFile(key string) (*multipart.FileHeader, error) {
	return c.Ctx.FormFile(key)
}

// Body returns the body of the request.
func (c *FiberContext) Body() ([]byte, error) {
	return c.Ctx.BodyRaw(), nil
}

// Method returns the HTTP method of the request.
func (c *FiberContext) Method() string {
	return c.Ctx.Method()
}

// Path returns the path of the request.
func (c *FiberContext) Path() string {
	return c.Ctx.Path()
}

// FullPath returns the route pattern of the request.
func (c *FiberContext) FullPath() string {
	if c.Ctx.Route().Path == "/" {
		return c.Ctx.Path()
	}

	return c.Ctx.Route().Path
}

// ClientIP returns the client IP of the request.
func (c *FiberContext) ClientIP() string {
	return c.Ctx.IP()
}

// Set sets the value of the context with the given key.
func (c *FiberContext) Set(key string, value any) {
	c.Ctx.Locals(key, value)
}

// Get returns the value of the context with the given key.
// It returns nil if the key does not exist.
func (c *FiberContext) Get(key string) (any, bool) {
	val := c.Ctx.Locals(key)
	if val != nil {
		return val, true
	}
	return val, false
}

// StatusCode returns the status code of the response.
func (c *FiberContext) StatusCode() int {
	return c.Ctx.Response().StatusCode()
}

// fiberResponseWriter is a response writer for Fiber.
type fiberResponseWriter struct {
	ctx       *fiber.Ctx
	header    http.Header
	wroteHead bool
}

// Header returns the header of the response.
func (w *fiberResponseWriter) Header() http.Header {
	return w.header
}

// WriteHeader writes the status code and header to the response.
func (w *fiberResponseWriter) WriteHeader(statusCode int) {
	if w.wroteHead {
		return
	}
	w.wroteHead = true

	for k, v := range w.header {
		for _, val := range v {
			w.ctx.Set(k, val)
		}
	}
	w.ctx.Status(statusCode)
}

// Write writes the data to the response.
func (w *fiberResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHead {
		w.WriteHeader(http.StatusOK)
	}
	return w.ctx.Write(b)
}

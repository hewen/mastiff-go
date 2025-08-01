// Package unicontext provides a context interface for HTTP handlers.
package unicontext

import (
	"mime/multipart"
	"net/http"
)

// UniversalContext is the interface for HTTP handlers.
type UniversalContext interface {
	// Request returns the HTTP request.
	Request() *http.Request
	// ResponseWriter returns the HTTP response writer.
	ResponseWriter() http.ResponseWriter
	// Next calls the next handler in the chain.
	Next() error

	// Param returns the value of the URL parameter with the given key.
	// It returns an empty string if the key does not exist.
	Param(key string) string
	// Query returns the value of the URL query parameter with the given key.
	// It returns an empty string if the key does not exist.
	Query(key string) string
	// Header returns the value of the HTTP header with the given key.
	// It returns an empty string if the key does not exist.
	Header(key string) string
	// Cookie returns the value of the HTTP cookie with the given key.
	// It returns an empty string if the key does not exist.
	Cookie(key string) string

	// JSON sends a JSON response with the given status code and data.
	JSON(status int, data any) error
	// Text sends a text response with the given status code and text.
	Text(status int, text string) error
	// String sends a string response with the given status code and formatted text.
	String(status int, format string, values ...any) error
	// HTML sends an HTML response with the given status code and HTML template name and data.
	HTML(status int, name string, obj any) error
	// Redirect sends a redirect response with the given status code and URL.
	Redirect(status int, url string) error

	// FormFile returns the file header of the form file with the given key.
	// It returns an error if the key does not exist.
	FormFile(key string) (*multipart.FileHeader, error)
	// File sends a file response with the given filepath.
	File(filepath string) error
	// Attachment sends an attachment response with the given filepath and filename.
	Attachment(filepath, filename string) error
	// BindJSON binds the JSON request body into the given object.
	BindJSON(target any) error
	// FormValue returns the value of the form field with the given key.
	// It returns an empty string if the key does not exist.
	FormValue(key string) string
	// Body returns the body of the request.
	Body() ([]byte, error)

	// Method returns the HTTP method of the request.
	Method() string
	// Path returns the path of the request.
	Path() string
	// FullPath returns the path of the request.
	FullPath() string
	// ClientIP returns the client IP of the request.
	ClientIP() string
	// Set sets the value of the context with the given key.
	Set(key string, value any)
	// Get returns the value of the context with the given key.
	// It returns nil if the key does not exist.
	Get(key string) (value any, exists bool)
	// StatusCode returns the status code of the response.
	StatusCode() int
}

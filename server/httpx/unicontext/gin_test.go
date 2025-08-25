package unicontext

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to run a single request test case.
func runGinRequestTest(t *testing.T, testCase struct {
	name           string
	method         string
	url            string
	body           string
	headers        map[string]string
	expectedMethod string
	expectedURL    string
	expectedBody   string
}) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Create request
	var bodyReader io.Reader
	if testCase.body != "" {
		bodyReader = strings.NewReader(testCase.body)
	}
	req := httptest.NewRequest(testCase.method, testCase.url, bodyReader)

	for key, value := range testCase.headers {
		req.Header.Set(key, value)
	}

	c.Request = req

	ginCtx := &GinContext{Ctx: c}
	request := ginCtx.Request()

	assert.Equal(t, testCase.expectedMethod, request.Method)
	assert.Equal(t, testCase.expectedURL, request.URL.String())

	if testCase.expectedBody != "" {
		body, err := io.ReadAll(request.Body)
		require.NoError(t, err)
		assert.Equal(t, testCase.expectedBody, string(body))
	}

	for key, expectedValue := range testCase.headers {
		assert.Equal(t, expectedValue, request.Header.Get(key))
	}
}

func TestGinContext_Request(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		method         string
		url            string
		body           string
		headers        map[string]string
		expectedMethod string
		expectedURL    string
		expectedBody   string
	}{
		{
			name:           "GET request",
			method:         "GET",
			url:            "/test?param=value",
			body:           "",
			headers:        map[string]string{"Content-Type": "application/json"},
			expectedMethod: "GET",
			expectedURL:    "/test?param=value",
			expectedBody:   "",
		},
		{
			name:           "POST request with body",
			method:         "POST",
			url:            "/api/users",
			body:           `{"name":"John"}`,
			headers:        map[string]string{"Content-Type": "application/json", "Authorization": "Bearer token"},
			expectedMethod: "POST",
			expectedURL:    "/api/users",
			expectedBody:   `{"name":"John"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runGinRequestTest(t, tt)
		})
	}
}

func TestGinContext_ResponseWriter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	ginCtx := &GinContext{Ctx: c}
	rw := ginCtx.ResponseWriter()

	assert.NotNil(t, rw)
	assert.Same(t, c.Writer, rw)
}

func TestGinContext_Next(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var nextCalled bool
	router := gin.New()
	router.Use(func(c *gin.Context) {
		ginCtx := &GinContext{Ctx: c}
		err := ginCtx.Next()
		assert.NoError(t, err)
	})

	router.GET("/test", func(c *gin.Context) {
		nextCalled = true
		c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.True(t, nextCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGinContext_Param(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/users/:id", func(c *gin.Context) {
		ginCtx := &GinContext{Ctx: c}
		id := ginCtx.Param("id")
		c.String(http.StatusOK, id)
	})

	req := httptest.NewRequest("GET", "/users/123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "123", w.Body.String())
}

func TestGinContext_Query(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/search", func(c *gin.Context) {
		ginCtx := &GinContext{Ctx: c}
		query := ginCtx.Query("q")
		c.String(http.StatusOK, query)
	})

	req := httptest.NewRequest("GET", "/search?q=golang", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "golang", w.Body.String())
}

func TestGinContext_Header(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/test", func(c *gin.Context) {
		ginCtx := &GinContext{Ctx: c}
		auth := ginCtx.Header("Authorization")
		c.String(http.StatusOK, auth)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer token123")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Bearer token123", w.Body.String())
}

func TestGinContext_Cookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/test", func(c *gin.Context) {
		ginCtx := &GinContext{Ctx: c}
		sessionID := ginCtx.Cookie("session_id")
		c.String(http.StatusOK, sessionID)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "abc123"})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "abc123", w.Body.String())
}

func TestGinContext_JSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	type Response struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	}

	router.POST("/test", func(c *gin.Context) {
		ginCtx := &GinContext{Ctx: c}
		data := Response{Message: "success", Code: 200}
		err := ginCtx.JSON(http.StatusOK, data)
		assert.NoError(t, err)
	})

	req := httptest.NewRequest("POST", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

	var result Response
	err := json.NewDecoder(w.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, "success", result.Message)
	assert.Equal(t, 200, result.Code)
}

func TestGinContext_Text(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/test", func(c *gin.Context) {
		ginCtx := &GinContext{Ctx: c}
		err := ginCtx.Text(http.StatusOK, "Hello World")
		assert.NoError(t, err)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Hello World", w.Body.String())
}

func TestGinContext_String(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/test", func(c *gin.Context) {
		ginCtx := &GinContext{Ctx: c}
		err := ginCtx.String(http.StatusOK, "Hello %s, you are %d years old", "John", 25)
		assert.NoError(t, err)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Hello John, you are 25 years old", w.Body.String())
}

func TestGinContext_HTML(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create test template directory and file first
	err := os.MkdirAll("testdata", 0750)
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll("testdata") }()

	templateContent := `<html><head><title>{{.title}}</title></head><body><h1>{{.title}}</h1></body></html>`
	err = os.WriteFile("testdata/test.html", []byte(templateContent), 0600)
	require.NoError(t, err)

	router := gin.New()
	// Load HTML templates for testing after creating the files
	router.LoadHTMLGlob("testdata/*.html")

	router.GET("/test", func(c *gin.Context) {
		ginCtx := &GinContext{Ctx: c}
		err := ginCtx.HTML(http.StatusOK, "test.html", gin.H{"title": "Test Page"})
		assert.NoError(t, err)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Test Page")
}

func TestGinContext_Redirect(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/test", func(c *gin.Context) {
		ginCtx := &GinContext{Ctx: c}
		err := ginCtx.Redirect(http.StatusFound, "/redirected")
		assert.NoError(t, err)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "/redirected", w.Header().Get("Location"))
}

func TestGinContext_File(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create a temporary file for testing
	content := "test file content"
	tmpFile := t.TempDir() + "/test.txt"
	err := os.WriteFile(tmpFile, []byte(content), 0600)
	require.NoError(t, err)

	router.GET("/test", func(c *gin.Context) {
		ginCtx := &GinContext{Ctx: c}
		err := ginCtx.File(tmpFile)
		assert.NoError(t, err)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, content, w.Body.String())
}

func TestGinContext_Attachment(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create a temporary file for testing
	content := "attachment content"
	tmpFile := t.TempDir() + "/attachment.txt"
	err := os.WriteFile(tmpFile, []byte(content), 0600)
	require.NoError(t, err)

	router.GET("/test", func(c *gin.Context) {
		ginCtx := &GinContext{Ctx: c}
		err := ginCtx.Attachment(tmpFile, "download.txt")
		assert.NoError(t, err)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Disposition"), "download.txt")
	assert.Equal(t, content, w.Body.String())
}

func TestGinContext_BindJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	type User struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	router.POST("/test", func(c *gin.Context) {
		ginCtx := &GinContext{Ctx: c}
		var user User
		err := ginCtx.BindJSON(&user)
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid JSON"})
			return
		}
		c.JSON(200, user)
	})

	userData := User{Name: "John Doe", Email: "john@example.com"}
	jsonData, _ := json.Marshal(userData)

	req := httptest.NewRequest("POST", "/test", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result User
	err := json.NewDecoder(w.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, userData.Name, result.Name)
	assert.Equal(t, userData.Email, result.Email)
}

func TestGinContext_FormValue(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.POST("/test", func(c *gin.Context) {
		ginCtx := &GinContext{Ctx: c}
		name := ginCtx.FormValue("name")
		c.String(http.StatusOK, name)
	})

	formData := "name=John+Doe&email=john%40example.com"
	req := httptest.NewRequest("POST", "/test", strings.NewReader(formData))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "John Doe", w.Body.String())
}

func TestGinContext_FormFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.POST("/test", func(c *gin.Context) {
		ginCtx := &GinContext{Ctx: c}
		file, err := ginCtx.FormFile("upload")
		if err != nil {
			c.String(400, "No file uploaded")
			return
		}
		c.String(http.StatusOK, file.Filename)
	})

	// Create multipart form data
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("upload", "test.txt")
	require.NoError(t, err)
	_, err = part.Write([]byte("file content"))
	require.NoError(t, err)
	_ = writer.Close()

	req := httptest.NewRequest("POST", "/test", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "test.txt", w.Body.String())
}

func TestGinContext_Method(t *testing.T) {
	gin.SetMode(gin.TestMode)
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			router := gin.New()

			router.Handle(method, "/test", func(c *gin.Context) {
				ginCtx := &GinContext{Ctx: c}
				c.String(http.StatusOK, ginCtx.Method())
			})

			req := httptest.NewRequest(method, "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, method, w.Body.String())
		})
	}
}

func TestGinContext_Path(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/users/123/posts", func(c *gin.Context) {
		ginCtx := &GinContext{Ctx: c}
		c.String(http.StatusOK, ginCtx.Path())
	})

	req := httptest.NewRequest("GET", "/users/123/posts", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "/users/123/posts", w.Body.String())
}

func TestGinContext_FullPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/users/:id/posts", func(c *gin.Context) {
		ginCtx := &GinContext{Ctx: c}
		c.String(http.StatusOK, ginCtx.FullPath())
	})

	req := httptest.NewRequest("GET", "/users/123/posts", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "/users/:id/posts", w.Body.String())
}

func TestGinContext_ClientIP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/test", func(c *gin.Context) {
		ginCtx := &GinContext{Ctx: c}
		c.String(http.StatusOK, ginCtx.ClientIP())
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.1")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// Gin should extract the IP from X-Forwarded-For header
	assert.NotEmpty(t, w.Body.String())
}

func TestGinContext_RemoteAddr(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/test", func(c *gin.Context) {
		ginCtx := &GinContext{Ctx: c}
		c.String(http.StatusOK, ginCtx.RemoteAddr())
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.1")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// Gin should extract the IP from X-Forwarded-For header
	assert.NotEmpty(t, w.Body.String())
}

func TestGinContext_SetAndGet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/test", func(c *gin.Context) {
		ginCtx := &GinContext{Ctx: c}

		// Test setting and getting a value
		ginCtx.Set("user_id", "123")
		ginCtx.Set("is_admin", true)

		userID, exists := ginCtx.Get("user_id")
		if !exists {
			c.String(500, "user_id not found")
			return
		}

		isAdmin, exists := ginCtx.Get("is_admin")
		if !exists {
			c.String(500, "is_admin not found")
			return
		}

		// Test getting non-existent key
		_, exists = ginCtx.Get("non_existent")
		if exists {
			c.String(500, "non_existent should not exist")
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"user_id":  userID,
			"is_admin": isAdmin,
		})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, "123", result["user_id"])
	assert.Equal(t, true, result["is_admin"])
}

func TestGinContext_StatusCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/test", func(c *gin.Context) {
		ginCtx := &GinContext{Ctx: c}
		c.Status(201)
		statusCode := ginCtx.StatusCode()
		c.String(http.StatusOK, fmt.Sprintf("%d", statusCode))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "201", w.Body.String())
}

func TestGinContext_InterfaceCompliance(t *testing.T) {
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		ginCtx := &GinContext{Ctx: c}
		_ = ginCtx.Method()
		c.Status(201)
		statusCode := ginCtx.StatusCode()
		c.String(http.StatusOK, fmt.Sprintf("%d", statusCode))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "201", w.Body.String())
}

// nolint
func TestGinContext_EdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Empty values", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)

		ginCtx := &GinContext{Ctx: c}

		// Test empty param
		assert.Equal(t, "", ginCtx.Param("nonexistent"))

		// Test empty query
		assert.Equal(t, "", ginCtx.Query("nonexistent"))

		// Test empty header
		assert.Equal(t, "", ginCtx.Header("nonexistent"))

		// Test empty cookie
		assert.Equal(t, "", ginCtx.Cookie("nonexistent"))

		// Test empty form value
		assert.Equal(t, "", ginCtx.FormValue("nonexistent"))
	})

	t.Run("Get non-existent key", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		ginCtx := &GinContext{Ctx: c}

		value, exists := ginCtx.Get("nonexistent")
		assert.Nil(t, value)
		assert.False(t, exists)
	})

	t.Run("BindJSON with invalid JSON", func(t *testing.T) {
		router := gin.New()

		router.POST("/test", func(c *gin.Context) {
			ginCtx := &GinContext{Ctx: c}
			var data map[string]interface{}
			err := ginCtx.BindJSON(&data)
			if err != nil {
				c.String(400, "Invalid JSON")
				return
			}
			c.JSON(200, data)
		})

		req := httptest.NewRequest("POST", "/test", strings.NewReader("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, 400, w.Code)
		assert.Equal(t, "Invalid JSON", w.Body.String())
	})

	t.Run("FormFile with no file", func(t *testing.T) {
		router := gin.New()

		router.POST("/test", func(c *gin.Context) {
			ginCtx := &GinContext{Ctx: c}
			_, err := ginCtx.FormFile("upload")
			if err != nil {
				c.String(400, "No file")
				return
			}
			c.String(200, "File found")
		})

		req := httptest.NewRequest("POST", "/test", strings.NewReader(""))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, 400, w.Code)
		assert.Equal(t, "No file", w.Body.String())
	})
}

func TestGinContext_Body(t *testing.T) {
	router := gin.New()
	router.POST("/test", func(c *gin.Context) {
		ginCtx := &GinContext{Ctx: c}
		body, _ := ginCtx.Body()
		_ = ginCtx.Data(http.StatusOK, "application/octet-stream", body)
	})

	data := "body"
	req := httptest.NewRequest("POST", "/test", strings.NewReader(data))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "body", w.Body.String())
}

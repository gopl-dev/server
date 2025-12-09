package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/app/service"
	"github.com/gopl-dev/server/server"
	"github.com/gopl-dev/server/server/request"
	"github.com/gopl-dev/server/server/response"
	"github.com/gopl-dev/server/test"
	"github.com/gopl-dev/server/test/factory"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/logrusorgru/aurora"
)

func init() {

}

type Test struct {
	Conf    *app.ConfigT
	DB      *pgxpool.Pool
	Service *service.Service
	Factory *factory.Factory
}

type Headers map[string]string

type Request struct {
	method       string
	path         string
	body         any
	bodyReader   io.Reader
	authBearer   string
	headers      Headers
	bindResponse any
	assertStatus int
}

var (
	authUser  *ds.User
	authToken string
	router    http.Handler
	tt        *Test
)

const ContentTypeJSON = "application/json"

func TestMain(m *testing.M) {
	tt = new(Test)
	ctx := context.Background()
	tt.Conf = app.Config()

	var err error

	tt.DB, err = app.NewPool(ctx)
	if err != nil {
		panic("TEST MAIN: " + err.Error())
	}

	err = app.MigrateDB(ctx, tt.DB)
	if err != nil {
		log.Fatal(err)
	}

	_, err = tt.DB.Exec(ctx, "BEGIN")
	if err != nil {
		println("[BEGIN TRANSACTION]:", err.Error())
		return
	}

	tt.Service = service.New(tt.DB)
	tt.Factory = factory.New(tt.DB)
	router = server.New(tt.Service).Handler
	code := m.Run()

	// shutdown
	_, err = tt.DB.Exec(ctx, "ROLLBACK")
	if err != nil {
		println("[SHUTDOWN] [ROLLBACK]:", err.Error())
	}

	tt.DB.Close()
	os.Exit(code)
}

func makeRequest(t *testing.T, r Request) *httptest.ResponseRecorder {
	t.Helper()
	req, err := http.NewRequestWithContext(
		context.Background(),
		r.method,
		r.path,
		r.bodyReader,
	)
	test.CheckErr(t, err)

	if r.method == http.MethodPost || r.method == http.MethodPut {
		req.Header.Set("Content-Type", ContentTypeJSON)
	}

	req.Header.Set("Accept", ContentTypeJSON)

	for k, v := range r.headers {
		req.Header.Set(k, v)
	}

	bearer := authToken
	if r.authBearer != "" {
		bearer = r.authBearer
	}

	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	return w
}

// Do sends given request,
// asserts that status is correct
// binds response and returns response.
func Do(t *testing.T, req Request) *httptest.ResponseRecorder {
	t.Helper()

	var (
		body []byte
		err  error
	)

	req.path = path.Join("/", tt.Conf.Server.APIBasePath, req.path)

	if !strings.HasSuffix(req.path, "/") && !strings.Contains(req.path, "/?") {
		req.path += "/"
	}

	if len(req.headers) == 0 {
		req.headers = Headers{}
	}

	if req.body != nil {
		body, err = json.MarshalIndent(req.body, "", "    ")
		test.CheckErr(t, err)
	}

	req.bodyReader = bytes.NewReader(body)

	resp := makeRequest(t, req)
	if resp.Code != req.assertStatus {
		t.Errorf(aurora.Red("(%d) %s %s").Bold().String(), resp.Code, req.method, req.path)

		if req.body != nil {
			println(aurora.Bold("Request:").String())
			println(string(body))
		}

		// try to bind error to struct and print it nicely
		// otherwise print it as is
		if responseContentType(resp) == ContentTypeJSON {
			var respBody map[string]any

			err = json.Unmarshal(resp.Body.Bytes(), &respBody)
			if err == nil {
				body2, err2 := json.MarshalIndent(respBody, "", "  ")
				if err2 != nil {
					t.Errorf("Response:\n%s", resp.Body.String())
				} else {
					t.Errorf("Response:\n%s", string(body2))
				}
			} else {
				t.Errorf("Response:\n%s", resp.Body.String())
			}
		}

		t.FailNow()
		return resp
	}

	t.Logf(aurora.Green("(%d) %s %s").Bold().String(), resp.Code, req.method, req.path)

	if responseContentType(resp) == ContentTypeJSON {
		err = json.Unmarshal(resp.Body.Bytes(), &req.bindResponse)
		if err != nil {
			t.Log(resp.Body.String())
			t.Fatal(err)
		}
	}

	return resp
}

func responseContentType(resp *httptest.ResponseRecorder) string {
	ct, ok := resp.Header()["Content-Type"]
	if !ok || len(ct) == 0 {
		return ""
	}

	frags := strings.Split(ct[0], ";")
	return frags[0]
}

// POST is a wrapper for Do.
func POST(t *testing.T, req Request) *httptest.ResponseRecorder {
	t.Helper()
	req.method = http.MethodPost
	return Do(t, req)
}

// PUT is a wrapper for Do.
func PUT(t *testing.T, req Request) *httptest.ResponseRecorder {
	t.Helper()

	req.method = http.MethodPut
	return Do(t, req)
}

// PATCH is a wrapper for Do.
func PATCH(t *testing.T, req Request) *httptest.ResponseRecorder {
	t.Helper()
	req.method = http.MethodPatch
	return Do(t, req)
}

// DELETE is a wrapper for Do.
func DELETE(t *testing.T, req Request) *httptest.ResponseRecorder {
	t.Helper()
	req.method = http.MethodDelete
	return Do(t, req)
}

// GET is a wrapper for Do.
func GET(t *testing.T, req Request) *httptest.ResponseRecorder {
	t.Helper()

	req.method = http.MethodGet
	return Do(t, req)
}

// testCREATE makes "create" POST request that expects response type of response arg and 201 status code.
func testCREATE(t *testing.T, path string, body, response any) *httptest.ResponseRecorder {
	t.Helper()

	req := Request{
		path:         path,
		body:         body,
		bindResponse: response,
		assertStatus: http.StatusCreated,
	}

	return POST(t, req)
}

// testUPDATE makes "update" request.
func testUPDATE(t *testing.T, path string, body, response any) *httptest.ResponseRecorder {
	t.Helper()

	req := Request{
		path:         path,
		body:         body,
		bindResponse: response,
		assertStatus: http.StatusOK,
	}

	return PUT(t, req)
}

// testDELETE makes "delete" request.
func testDELETE(t *testing.T, path string, response any) *httptest.ResponseRecorder {
	t.Helper()
	req := Request{
		path:         path,
		bindResponse: response,
		assertStatus: http.StatusOK,
	}

	return DELETE(t, req)
}

// testGET makes simple "get" request.
func testGET(t *testing.T, path string, response any) *httptest.ResponseRecorder {
	t.Helper()

	req := Request{
		path:         path,
		bindResponse: response,
		assertStatus: http.StatusOK,
	}

	return GET(t, req)
}

// testQuery makes "get" request with query params.
func testQuery(t *testing.T, path string, request, response any) *httptest.ResponseRecorder {
	t.Helper()
	query, err := test.StructToQueryString(request)
	test.CheckErr(t, err)

	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	if query != "" {
		path += "?" + query
	}

	req := Request{
		path:         path,
		bindResponse: response,
		assertStatus: http.StatusOK,
	}

	return GET(t, req)
}

func login(t *testing.T) *ds.User {
	t.Helper()

	if authUser != nil && authToken != "" {
		return authUser
	}

	// create new user and make auth request to get auth token
	password := "mytestpassword"
	usr := tt.Factory.CreateUser(t, ds.User{
		Password: password,
	})

	resp := response.UserSignIn{}
	POST(t, Request{
		path:         "auth",
		body:         request.UserSignIn{Email: usr.Email, Password: password},
		bindResponse: &resp,
		assertStatus: 200,
	})

	authToken = resp.Token
	authUser = usr

	return authUser
}

// Shortcut to fmt.Sprint
// Hey, this is not laziness, this is speed ^^.
func f(s string, args ...any) string {
	return fmt.Sprintf(s, args...)
}

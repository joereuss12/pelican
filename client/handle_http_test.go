//go:build !windows

/***************************************************************
 *
 * Copyright (C) 2024, University of Nebraska-Lincoln
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you
 * may not use this file except in compliance with the License.  You may
 * obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 ***************************************************************/

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pelicanplatform/pelican/config"
	"github.com/pelicanplatform/pelican/mock"
	"github.com/pelicanplatform/pelican/namespaces"
	"github.com/pelicanplatform/pelican/server_utils"
	"github.com/pelicanplatform/pelican/test_utils"
)

func TestMain(m *testing.M) {
	viper.Reset()
	if err := config.InitClient(); err != nil {
		os.Exit(1)
	}
	os.Exit(m.Run())
}

// TestIsPort calls main.hasPort with a hostname, checking
// for a valid return value.
func TestIsPort(t *testing.T) {

	if hasPort("blah.not.port:") {
		t.Fatal("Failed to parse port when : at end")
	}

	if !hasPort("host:1") {
		t.Fatal("Failed to parse with port = 1")
	}

	if hasPort("https://example.com") {
		t.Fatal("Failed when scheme is specified")
	}
}

// TestNewTransferDetails checks the creation of transfer details
func TestNewTransferDetails(t *testing.T) {
	os.Setenv("http_proxy", "http://proxy.edu:3128")
	t.Cleanup(func() {
		require.NoError(t, os.Unsetenv("http_proxy"))
	})

	// Case 1: cache with http
	testCache := namespaces.Cache{
		AuthEndpoint: "cache.edu:8443",
		Endpoint:     "cache.edu:8000",
		Resource:     "Cache",
	}
	transfers := newTransferDetails(testCache, transferDetailsOptions{false, ""})
	assert.Equal(t, 2, len(transfers))
	assert.Equal(t, "cache.edu:8000", transfers[0].Url.Host)
	assert.Equal(t, "http", transfers[0].Url.Scheme)
	assert.Equal(t, true, transfers[0].Proxy)
	assert.Equal(t, "cache.edu:8000", transfers[1].Url.Host)
	assert.Equal(t, "http", transfers[1].Url.Scheme)
	assert.Equal(t, false, transfers[1].Proxy)

	// Case 2: cache with https
	transfers = newTransferDetails(testCache, transferDetailsOptions{true, ""})
	assert.Equal(t, 1, len(transfers))
	assert.Equal(t, "cache.edu:8443", transfers[0].Url.Host)
	assert.Equal(t, "https", transfers[0].Url.Scheme)
	assert.Equal(t, false, transfers[0].Proxy)

	testCache.Endpoint = "cache.edu"
	// Case 3: cache without port with http
	transfers = newTransferDetails(testCache, transferDetailsOptions{false, ""})
	assert.Equal(t, 2, len(transfers))
	assert.Equal(t, "cache.edu:8000", transfers[0].Url.Host)
	assert.Equal(t, "http", transfers[0].Url.Scheme)
	assert.Equal(t, true, transfers[0].Proxy)
	assert.Equal(t, "cache.edu:8000", transfers[1].Url.Host)
	assert.Equal(t, "http", transfers[1].Url.Scheme)
	assert.Equal(t, false, transfers[1].Proxy)

	// Case 4. cache without port with https
	testCache.AuthEndpoint = "cache.edu"
	transfers = newTransferDetails(testCache, transferDetailsOptions{true, ""})
	assert.Equal(t, 2, len(transfers))
	assert.Equal(t, "cache.edu:8444", transfers[0].Url.Host)
	assert.Equal(t, "https", transfers[0].Url.Scheme)
	assert.Equal(t, false, transfers[0].Proxy)
	assert.Equal(t, "cache.edu:8443", transfers[1].Url.Host)
	assert.Equal(t, "https", transfers[1].Url.Scheme)
	assert.Equal(t, false, transfers[1].Proxy)
}

func TestNewTransferDetailsEnv(t *testing.T) {
	os.Setenv("http_proxy", "http://proxy.edu:3128")
	t.Cleanup(func() {
		require.NoError(t, os.Unsetenv("http_proxy"))
	})

	testCache := namespaces.Cache{
		AuthEndpoint: "cache.edu:8443",
		Endpoint:     "cache.edu:8000",
		Resource:     "Cache",
	}

	os.Setenv("OSG_DISABLE_PROXY_FALLBACK", "")
	err := config.InitClient()
	assert.Nil(t, err)
	transfers := newTransferDetails(testCache, transferDetailsOptions{})
	assert.Equal(t, 1, len(transfers))
	assert.Equal(t, true, transfers[0].Proxy)

	os.Unsetenv("http_proxy")

	transfers = newTransferDetails(testCache, transferDetailsOptions{true, ""})
	assert.Equal(t, 1, len(transfers))
	assert.Equal(t, "https", transfers[0].Url.Scheme)
	assert.Equal(t, false, transfers[0].Proxy)
	os.Unsetenv("OSG_DISABLE_PROXY_FALLBACK")
	viper.Reset()
	err = config.InitClient()
	assert.Nil(t, err)
}

func TestSlowTransfers(t *testing.T) {
	ctx, _, _ := test_utils.TestContext(context.Background(), t)

	// Adjust down some timeouts to speed up the test
	viper.Set("Client.SlowTransferWindow", 5)
	viper.Set("Client.SlowTransferRampupTime", 10)

	channel := make(chan bool)
	slowDownload := 1024 * 10 // 10 KiB/s < 100 KiB/s
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "HEAD" {
			w.Header().Add("Content-Length", "1024000")
			w.WriteHeader(http.StatusOK)
			return
		}
		buffer := make([]byte, slowDownload)
		for {
			select {
			case <-channel:
				return
			default:
				_, err := w.Write(buffer)
				if err != nil {
					return
				}
				w.(http.Flusher).Flush()
				time.Sleep(1 * time.Second)
			}
		}
	}))

	defer svr.CloseClientConnections()
	defer svr.Close()

	testCache := namespaces.Cache{
		AuthEndpoint: svr.URL,
		Endpoint:     svr.URL,
		Resource:     "Cache",
	}

	os.Setenv("http_proxy", "http://proxy.edu:3128")
	t.Cleanup(func() {
		require.NoError(t, os.Unsetenv("http_proxy"))
	})

	transfers := newTransferDetails(testCache, transferDetailsOptions{false, ""})
	assert.Equal(t, 2, len(transfers))
	assert.Equal(t, svr.URL, transfers[0].Url.String())

	finishedChannel := make(chan bool)
	var err error
	// Do a quick timeout
	go func() {
		_, _, _, err = downloadHTTP(ctx, nil, nil, transfers[0], filepath.Join(t.TempDir(), "test.txt"), -1, "", "")
		finishedChannel <- true
	}()

	select {
	case <-finishedChannel:
		if err == nil {
			t.Fatal("Error is nil, download should have failed")
		}
	case <-time.After(time.Second * 160):
		// 120 seconds for warmup, 30 seconds for download
		t.Fatal("Maximum downloading time reach, download should have failed")
	}

	// Close the channel to allow the download to complete
	channel <- true

	// Make sure the errors are correct
	assert.NotNil(t, err)
	assert.IsType(t, &SlowTransferError{}, err)
}

// Test stopped transfer
func TestStoppedTransfer(t *testing.T) {
	os.Setenv("http_proxy", "http://proxy.edu:3128")
	t.Cleanup(func() {
		require.NoError(t, os.Unsetenv("http_proxy"))
	})

	ctx, _, _ := test_utils.TestContext(context.Background(), t)

	// Adjust down the timeouts
	viper.Set("Client.StoppedTransferTimeout", 3)
	viper.Set("Client.SlowTransferRampupTime", 100)

	channel := make(chan bool)
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "HEAD" {
			w.Header().Add("Content-Length", "102400")
			w.WriteHeader(http.StatusOK)
			return
		}
		buffer := make([]byte, 1024*100)
		for {
			select {
			case <-channel:
				return
			default:
				_, err := w.Write(buffer)
				if err != nil {
					return
				}
				w.(http.Flusher).Flush()
				time.Sleep(1 * time.Second)
				buffer = make([]byte, 0)
			}
		}
	}))

	defer svr.CloseClientConnections()
	defer svr.Close()

	testCache := namespaces.Cache{
		AuthEndpoint: svr.URL,
		Endpoint:     svr.URL,
		Resource:     "Cache",
	}
	transfers := newTransferDetails(testCache, transferDetailsOptions{false, ""})
	assert.Equal(t, 2, len(transfers))
	assert.Equal(t, svr.URL, transfers[0].Url.String())

	finishedChannel := make(chan bool)
	var err error

	go func() {
		_, _, _, err = downloadHTTP(ctx, nil, nil, transfers[0], filepath.Join(t.TempDir(), "test.txt"), -1, "", "")
		finishedChannel <- true
	}()

	select {
	case <-finishedChannel:
		if err == nil {
			t.Fatal("Download should have failed")
		}
	case <-time.After(time.Second * 150):
		t.Fatal("Download should have failed")
	}

	// Close the channel to allow the download to complete
	channel <- true

	// Make sure the errors are correct
	assert.NotNil(t, err)
	assert.IsType(t, &StoppedTransferError{}, err, err.Error())
	assert.True(t, IsRetryable(err))
}

// Test connection error
func TestConnectionError(t *testing.T) {
	ctx, _, _ := test_utils.TestContext(context.Background(), t)

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("dialClosedPort: Listen failed: %v", err)
	}
	addr := l.Addr().String()
	l.Close()

	_, _, _, err = downloadHTTP(ctx, nil, nil, transferAttemptDetails{Url: &url.URL{Host: addr, Scheme: "http"}, Proxy: false}, filepath.Join(t.TempDir(), "test.txt"), -1, "", "")

	assert.IsType(t, &ConnectionSetupError{}, err)

}

func TestTrailerError(t *testing.T) {
	ctx, _, _ := test_utils.TestContext(context.Background(), t)

	// Set up an HTTP server that returns an error trailer
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Trailer", "X-Transfer-Status")
		w.Header().Set("X-Transfer-Status", "500: Unable to read test.txt; input/output error")

		chunkedWriter := httputil.NewChunkedWriter(w)
		defer chunkedWriter.Close()

		_, err := chunkedWriter.Write([]byte("Test data"))
		if err != nil {
			t.Fatalf("Error writing to chunked writer: %v", err)
		}
	}))

	defer svr.Close()

	os.Setenv("http_proxy", "http://proxy.edu:3128")
	t.Cleanup(func() {
		require.NoError(t, os.Unsetenv("http_proxy"))
	})

	testCache := namespaces.Cache{
		AuthEndpoint: svr.URL,
		Endpoint:     svr.URL,
		Resource:     "Cache",
	}
	transfers := newTransferDetails(testCache, transferDetailsOptions{false, ""})
	assert.Equal(t, 2, len(transfers))
	assert.Equal(t, svr.URL, transfers[0].Url.String())

	// Call DownloadHTTP and check if the error is returned correctly
	_, _, _, err := downloadHTTP(ctx, nil, nil, transfers[0], filepath.Join(t.TempDir(), "test.txt"), -1, "", "")

	assert.NotNil(t, err)
	assert.EqualError(t, err, "transfer error: Unable to read test.txt; input/output error")
}

func TestUploadZeroLengthFile(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//t.Logf("%s", dump)
		assert.Equal(t, "PUT", r.Method, "Not PUT Method")
		assert.Equal(t, int64(0), r.ContentLength, "ContentLength should be 0")
	}))
	defer ts.Close()
	reader := bytes.NewReader([]byte{})
	request, err := http.NewRequest("PUT", ts.URL, reader)
	if err != nil {
		assert.NoError(t, err)
	}

	request.Header.Set("Authorization", "Bearer test")
	errorChan := make(chan error, 1)
	responseChan := make(chan *http.Response)
	go runPut(request, responseChan, errorChan)
	select {
	case err := <-errorChan:
		assert.NoError(t, err)
	case response := <-responseChan:
		assert.Equal(t, http.StatusOK, response.StatusCode)
	case <-time.After(time.Second * 2):
		assert.Fail(t, "Timeout while waiting for response")
	}
}

func TestFailedUpload(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//t.Logf("%s", dump)
		assert.Equal(t, "PUT", r.Method, "Not PUT Method")
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("Error"))
		assert.NoError(t, err)

	}))
	defer ts.Close()
	reader := strings.NewReader("test")
	request, err := http.NewRequest("PUT", ts.URL, reader)
	if err != nil {
		assert.NoError(t, err)
	}
	request.Header.Set("Authorization", "Bearer test")
	errorChan := make(chan error, 1)
	responseChan := make(chan *http.Response)
	go runPut(request, responseChan, errorChan)
	select {
	case err := <-errorChan:
		assert.Error(t, err)
	case response := <-responseChan:
		assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
	case <-time.After(time.Second * 2):
		assert.Fail(t, "Timeout while waiting for response")
	}
}

func TestSortAttempts(t *testing.T) {
	ctx, cancel, _ := test_utils.TestContext(context.Background(), t)

	neverRespond := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		select {
		case <-ctx.Done():
		case <-ticker.C:
		}
	})
	alwaysRespond := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "HEAD" {
			w.Header().Set("Content-Length", "42")
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
	svr1 := httptest.NewServer(neverRespond)
	defer svr1.Close()
	url1, err := url.Parse(svr1.URL)
	require.NoError(t, err)
	attempt1 := transferAttemptDetails{Url: url1}

	svr2 := httptest.NewServer(alwaysRespond)
	defer svr2.Close()
	url2, err := url.Parse(svr2.URL)
	require.NoError(t, err)
	attempt2 := transferAttemptDetails{Url: url2}

	svr3 := httptest.NewServer(alwaysRespond)
	defer svr3.Close()
	url3, err := url.Parse(svr3.URL)
	require.NoError(t, err)
	attempt3 := transferAttemptDetails{Url: url3}

	defer cancel()

	size, results := sortAttempts(ctx, "/path", []transferAttemptDetails{attempt1, attempt2, attempt3})
	assert.Equal(t, int64(42), size)
	assert.Equal(t, svr2.URL, results[0].Url.String())
	assert.Equal(t, svr3.URL, results[1].Url.String())
	assert.Equal(t, svr1.URL, results[2].Url.String())

	size, results = sortAttempts(ctx, "/path", []transferAttemptDetails{attempt2, attempt3, attempt1})
	assert.Equal(t, int64(42), size)
	assert.Equal(t, svr2.URL, results[0].Url.String())
	assert.Equal(t, svr3.URL, results[1].Url.String())
	assert.Equal(t, svr1.URL, results[2].Url.String())

	size, results = sortAttempts(ctx, "/path", []transferAttemptDetails{attempt1, attempt1})
	assert.Equal(t, int64(-1), size)
	assert.Equal(t, svr1.URL, results[0].Url.String())
	assert.Equal(t, svr1.URL, results[1].Url.String())

	size, results = sortAttempts(ctx, "/path", []transferAttemptDetails{attempt2, attempt3})
	assert.Equal(t, int64(42), size)
	assert.Equal(t, svr2.URL, results[0].Url.String())
	assert.Equal(t, svr3.URL, results[1].Url.String())
}

func TestTimeoutHeaderSetForDownload(t *testing.T) {
	viper.Reset()
	config.InitConfig()
	viper.Set("Transport.ResponseHeaderTimeout", 10*time.Second)
	require.NoError(t, config.InitClient())
	ctx, _, _ := test_utils.TestContext(context.Background(), t)

	// We have this flag because our server will get a few requests throughout its lifetime and the other
	// requests do not contain the X-Pelican-Timeout header
	timeoutHeaderFound := false

	// Create a mock server to download from
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the "X-Pelican-Timeout" header is set
		if !timeoutHeaderFound {
			if r.Header.Get("X-Pelican-Timeout") == "" {
				t.Error("X-Pelican-Timeout header is not set")
			}
			assert.Equal(t, "9.5s", r.Header.Get("X-Pelican-Timeout"))
			timeoutHeaderFound = true
		}
	}))
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	assert.NoError(t, err)
	_, _, _, err = downloadHTTP(ctx, nil, nil, transferAttemptDetails{Url: serverURL, Proxy: false}, filepath.Join(t.TempDir(), "test.txt"), -1, "", "")
	assert.NoError(t, err)
	viper.Reset()
}

func TestJobIdHeaderSetForDownload(t *testing.T) {
	viper.Reset()
	config.InitConfig()
	require.NoError(t, config.InitClient())

	// Create a test .job.ad file
	jobAdFile, err := os.CreateTemp("", ".job.ad")
	assert.NoError(t, err)

	// Write the job ad to the file
	_, err = jobAdFile.WriteString("GlobalJobId = 12345")
	assert.NoError(t, err)
	jobAdFile.Close()

	os.Setenv("_CONDOR_JOB_AD", jobAdFile.Name())
	ctx, _, _ := test_utils.TestContext(context.Background(), t)

	// We have this flag because our server will get a few requests throughout its lifetime and the other
	// requests do not contain the X-Pelican-Timeout header
	timeoutHeaderFound := false

	// Create a mock server to download from
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the "X-Pelican-Timeout" header is set
		if !timeoutHeaderFound {
			if r.Header.Get("X-Pelican-JobId") == "" {
				t.Error("X-Pelican-JobId header is not set")
			}
			assert.Equal(t, "12345", r.Header.Get("X-Pelican-JobId"))
			timeoutHeaderFound = true
		}
	}))
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	assert.NoError(t, err)
	_, _, _, err = downloadHTTP(ctx, nil, nil, transferAttemptDetails{Url: serverURL, Proxy: false}, filepath.Join(t.TempDir(), "test.txt"), -1, "", "")
	assert.NoError(t, err)
	viper.Reset()
	os.Unsetenv("_CONDOR_JOB_AD")
}

// Server test object for testing user agent
type (
	server_test struct {
		server     *httptest.Server
		user_agent *string
	}
)

// Test to ensure the user-agent header is being updating in the request made within DownloadHTTP()
func TestProjInUserAgent(t *testing.T) {
	ctx, _, _ := test_utils.TestContext(context.Background(), t)

	server_test := server_test{}
	// Create a mock server to download from
	server_test.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Note: we check for this HEAD request because within DownloadHTTP() we make a HEAD request to get the content length
		// This request is a different user-agent header (and different request) so we need to ignore it so server_test.user_agent is not overwritten
		if r.Method == "HEAD" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		userAgent := r.UserAgent()
		server_test.user_agent = &userAgent
	}))
	defer server_test.server.Close()
	defer server_test.server.CloseClientConnections()

	serverURL, err := url.Parse(server_test.server.URL)
	assert.NoError(t, err)
	_, _, _, err = downloadHTTP(ctx, nil, nil, transferAttemptDetails{Url: serverURL, Proxy: false}, filepath.Join(t.TempDir(), "test.txt"), -1, "", "test")
	assert.NoError(t, err)

	// Test the user-agent header is what we expect it to be
	assert.Equal(t, "pelican-client/"+config.GetVersion()+" project/test", *server_test.user_agent)
}

func TestNewPelicanURL(t *testing.T) {
	// Set up our federation and context
	ctx, cancel, egrp := test_utils.TestContext(context.Background(), t)
	config.InitConfig()

	t.Run("TestOsdfOrStashSchemeWithOSDFPrefixNoError", func(t *testing.T) {
		viper.Reset()
		err := config.InitClient()
		require.NoError(t, err)
		_, err = config.SetPreferredPrefix(config.OsdfPrefix)
		viper.Set("ConfigDir", t.TempDir())
		assert.NoError(t, err)
		// Init config to get proper timeouts
		config.InitConfig()

		te, err := NewTransferEngine(ctx)
		require.NoError(t, err)

		defer func() {
			require.NoError(t, te.Shutdown())
		}()

		remoteObject := "osdf:///something/somewhere/thatdoesnotexist.txt"
		remoteObjectURL, err := url.Parse(remoteObject)
		assert.NoError(t, err)

		// Instead of relying on osdf, let's just set our global metadata (osdf prefix does this for us)
		viper.Set("Federation.DirectorUrl", "someDirectorUrl")
		viper.Set("Federation.DiscoveryUrl", "someDiscoveryUrl")

		pelicanURL, err := te.newPelicanURL(remoteObjectURL)
		assert.NoError(t, err)

		// Check pelicanURL properly filled out
		assert.Equal(t, "someDirectorUrl", pelicanURL.directorUrl)
		viper.Reset()
	})

	t.Run("TestOsdfOrStashSchemeWithOSDFPrefixWithError", func(t *testing.T) {
		viper.Reset()
		err := config.InitClient()
		require.NoError(t, err)
		_, err = config.SetPreferredPrefix(config.OsdfPrefix)
		viper.Set("ConfigDir", t.TempDir())
		require.NoError(t, err)
		config.InitConfig()
		require.NoError(t, config.InitClient())

		te, err := NewTransferEngine(ctx)
		require.NoError(t, err)
		defer func() {
			require.NoError(t, te.Shutdown())
		}()

		remoteObject := "osdf:///something/somewhere/thatdoesnotexist.txt"
		remoteObjectURL, err := url.Parse(remoteObject)
		assert.NoError(t, err)

		// Instead of relying on osdf, let's just set our global metadata but don't set one piece
		viper.Set("Federation.DiscoveryUrl", "someDiscoveryUrl")

		_, err = te.newPelicanURL(remoteObjectURL)
		// Make sure we get an error
		assert.Error(t, err)
		viper.Reset()
	})

	t.Run("TestOsdfOrStashSchemeWithPelicanPrefixNoError", func(t *testing.T) {
		viper.Reset()
		err := config.InitClient()
		require.NoError(t, err)
		viper.Set("ConfigDir", t.TempDir())
		config.InitConfig()
		require.NoError(t, config.InitClient())
		te, err := NewTransferEngine(ctx)
		require.NoError(t, err)
		defer func() {
			require.NoError(t, te.Shutdown())
		}()

		mock.MockOSDFDiscovery(t, config.GetTransport())
		_, err = config.SetPreferredPrefix(config.PelicanPrefix)
		config.InitConfig()
		assert.NoError(t, err)
		remoteObject := "osdf:///something/somewhere/thatdoesnotexist.txt"
		remoteObjectURL, err := url.Parse(remoteObject)
		assert.NoError(t, err)

		pelicanURL, err := te.newPelicanURL(remoteObjectURL)
		assert.NoError(t, err)

		// Check pelicanURL properly filled out
		assert.Equal(t, "https://osdf-director.osg-htc.org", pelicanURL.directorUrl)
		viper.Reset()
		// Note: can't really test this for an error since that would require osg-htc.org to be down
	})

	t.Run("TestPelicanSchemeNoError", func(t *testing.T) {
		viper.Reset()
		viper.Set("TLSSkipVerify", true)
		viper.Set("ConfigDir", t.TempDir())
		config.InitConfig()
		err := config.InitClient()
		require.NoError(t, err)

		te, err := NewTransferEngine(ctx)
		require.NoError(t, err)
		defer func() {
			require.NoError(t, te.Shutdown())
		}()

		assert.NoError(t, err)
		// Create a server that gives us a mock response
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// make our response:
			response := config.FederationDiscovery{
				DirectorEndpoint:              "director",
				NamespaceRegistrationEndpoint: "registry",
				JwksUri:                       "jwks",
				BrokerEndpoint:                "broker",
			}

			responseJSON, err := json.Marshal(response)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			_, err = w.Write(responseJSON)
			assert.NoError(t, err)
		}))
		defer server.Close()

		serverURL, err := url.Parse(server.URL)
		assert.NoError(t, err)

		remoteObject := "pelican://" + serverURL.Host + "/something/somewhere/thatdoesnotexist.txt"
		remoteObjectURL, err := url.Parse(remoteObject)
		assert.NoError(t, err)

		pelicanURL, err := te.newPelicanURL(remoteObjectURL)
		assert.NoError(t, err)

		// Check pelicanURL properly filled out
		assert.Equal(t, "director", pelicanURL.directorUrl)
		// Check to make sure it was populated in our cache
		assert.True(t, te.pelicanURLCache.Has("https://"+serverURL.Host))
		viper.Reset()
	})

	t.Run("TestPelicanSchemeWithError", func(t *testing.T) {
		viper.Reset()
		viper.Set("ConfigDir", t.TempDir())
		config.InitConfig()
		err := config.InitClient()
		require.NoError(t, err)

		te, err := NewTransferEngine(ctx)
		require.NoError(t, err)
		defer func() {
			require.NoError(t, te.Shutdown())
		}()

		remoteObject := "pelican://some-host/something/somewhere/thatdoesnotexist.txt"
		remoteObjectURL, err := url.Parse(remoteObject)
		assert.NoError(t, err)

		_, err = te.newPelicanURL(remoteObjectURL)
		assert.Error(t, err)
		viper.Reset()
	})

	t.Run("TestPelicanSchemeMetadataTimeoutError", func(t *testing.T) {
		viper.Reset()
		viper.Set("TLSSkipVerify", true)
		viper.Set("ConfigDir", t.TempDir())
		oldResponseHeaderTimeout := viper.Get("transport.ResponseHeaderTimeout")
		viper.Set("transport.ResponseHeaderTimeout", 0.1*float64(time.Millisecond))
		viper.Set("Client.WorkerCount", 5)
		err := config.InitClient()
		require.NoError(t, err)

		te, err := NewTransferEngine(ctx)
		require.NoError(t, err)
		defer func() {
			require.NoError(t, te.Shutdown())
		}()

		// Create a server that gives us a mock response
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// make our response:
			response := config.FederationDiscovery{
				DirectorEndpoint:              "director",
				NamespaceRegistrationEndpoint: "registry",
				JwksUri:                       "jwks",
				BrokerEndpoint:                "broker",
			}

			responseJSON, err := json.Marshal(response)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			_, err = w.Write(responseJSON)
			assert.NoError(t, err)
		}))
		defer server.Close()

		serverURL, err := url.Parse(server.URL)
		assert.NoError(t, err)

		remoteObject := "pelican://" + serverURL.Host + "/something/somewhere/thatdoesnotexist.txt"
		remoteObjectURL, err := url.Parse(remoteObject)
		assert.NoError(t, err)

		_, err = te.newPelicanURL(remoteObjectURL)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, config.MetadataTimeoutErr))
		viper.Set("transport.ResponseHeaderTimeout", oldResponseHeaderTimeout)
	})

	t.Cleanup(func() {
		cancel()
		if err := egrp.Wait(); err != nil && err != context.Canceled && err != http.ErrServerClosed {
			require.NoError(t, err)
		}
		// Throw in a viper.Reset for good measure. Keeps our env squeaky clean!
		viper.Reset()
	})
}

// Tests the functionality of getCachesToTry, ensuring that the function returns the correct number of caches and removes duplicates
func TestGetCachesToTry(t *testing.T) {
	directorCaches := make([]namespaces.DirectorCache, 3)
	for i := 0; i < 3; i++ {
		directorCache := namespaces.DirectorCache{
			EndpointUrl: "https://some/cache/" + strconv.Itoa(i),
			Priority:    0,
			AuthedReq:   false,
		}
		directorCaches[i] = directorCache
	}

	// Add a duplicate to the list --> check for its removal
	directorCaches = append(directorCaches, namespaces.DirectorCache{
		EndpointUrl: "https://some/cache/0",
		Priority:    0,
		AuthedReq:   false,
	})

	// Make our namespace:
	namespace := namespaces.Namespace{
		SortedDirectorCaches: directorCaches,
		ReadHTTPS:            false,
		UseTokenOnRead:       false,
	}

	caches, err := getCachesFromNamespace(namespace, true, nil)
	assert.NoError(t, err)

	job := &TransferJob{
		namespace: namespace,
	}

	transfers := getCachesToTry(caches, job, 4, "")

	// Check that there are no duplicates in the result
	cacheSet := make(map[CacheInterface]bool)
	for _, transfer := range transfers {
		if cacheSet[transfer.Url.String()] {
			t.Errorf("Found duplicate cache: %v", transfer.Url.String())
		}
		cacheSet[transfer.Url.String()] = true
	}
	// Verify we got the correct caches in our transfer attempt details
	require.Len(t, transfers, 3)
	assert.Equal(t, "https://some/cache/0", transfers[0].Url.String())
	assert.Equal(t, "https://some/cache/1", transfers[1].Url.String())
	assert.Equal(t, "https://some/cache/2", transfers[2].Url.String())
}

// Test that the project name is correctly extracted from the job ad file
func TestSearchJobAd(t *testing.T) {
	// Create a temporary file
	tempFile, err := os.CreateTemp("", "test")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	// Write a project name and job id to the file
	_, err = tempFile.WriteString("ProjectName = \"testProject\"\nGlobalJobId = 12345")
	assert.NoError(t, err)
	tempFile.Close()
	t.Run("TestNoJobAd", func(t *testing.T) {
		// Unset this environment var
		os.Unsetenv("_CONDOR_JOB_AD")
		// Call GetProjectName and check the result
		projectName := searchJobAd(projectName)
		assert.Equal(t, "", projectName)
	})

	t.Run("TestProjectNameAd", func(t *testing.T) {
		// Set the _CONDOR_JOB_AD environment variable to the temp file's name
		os.Setenv("_CONDOR_JOB_AD", tempFile.Name())
		defer os.Unsetenv("_CONDOR_JOB_AD")

		// Call GetProjectName and check the result
		projectName := searchJobAd(projectName)
		assert.Equal(t, "testProject", projectName)
	})

	t.Run("TestGlobalJobIdAd", func(t *testing.T) {
		// Set the _CONDOR_JOB_AD environment variable to the temp file's name
		os.Setenv("_CONDOR_JOB_AD", tempFile.Name())
		defer os.Unsetenv("_CONDOR_JOB_AD")

		// Call GetProjectName and check the result
		jobId := searchJobAd(jobId)
		assert.Equal(t, "12345", jobId)
	})
}

func TestNewTransferEngine(t *testing.T) {
	viper.Reset()
	defer viper.Reset()
	// Test we fail if we do not call initclient() before
	t.Run("TestInitClientNotCalled", func(t *testing.T) {
		config.ResetClientInitialized()
		ctx := context.Background()
		_, err := NewTransferEngine(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "client has not been initialized, unable to create transfer engine")
	})

	t.Run("TestInitClientCalled", func(t *testing.T) {
		err := config.InitClient()
		require.NoError(t, err)
		ctx := context.Background()
		_, err = NewTransferEngine(ctx)
		assert.NoError(t, err)
	})
}

func TestGetDirListHost(t *testing.T) {
	// Test we get dirlisthost with valid PROPFIND on test server
	t.Run("testValidPropfind", func(t *testing.T) {
		expectedLocation := "http://some/origin/path/to/object"
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Location", expectedLocation)
			w.WriteHeader(http.StatusTemporaryRedirect)
		}
		server := httptest.NewServer(http.HandlerFunc(handler))
		defer server.Close()
		ctx := context.Background()
		testObjectUrl, err := url.Parse("pelican://federation/some/object")
		require.NoError(t, err)

		// Wait until our server is working
		err = server_utils.WaitUntilWorking(ctx, "PROPFIND", server.URL, "testServer", http.StatusTemporaryRedirect, true)
		require.NoError(t, err)

		dirListHost, err := getDirListHost(ctx, testObjectUrl, namespaces.Namespace{}, server.URL)
		require.NoError(t, err)
		assert.Equal(t, "http://some", dirListHost.String())
	})

	// Test we get dirlist host when PROPFIND returns 405 but dirlisthost set in namespace
	t.Run("testInvalidPropfindValidDirListInNamespace", func(t *testing.T) {
		expectedLocation := "http://origin"
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		server := httptest.NewServer(http.HandlerFunc(handler))
		defer server.Close()
		ctx := context.Background()
		testObjectUrl, err := url.Parse("pelican://federation/some/object")
		require.NoError(t, err)

		// Wait until our server is working
		err = server_utils.WaitUntilWorking(ctx, "PROPFIND", server.URL, "testServer", http.StatusMethodNotAllowed, true)
		require.NoError(t, err)

		dirListHost, err := getDirListHost(ctx, testObjectUrl, namespaces.Namespace{DirListHost: expectedLocation}, server.URL)
		require.NoError(t, err)
		assert.Equal(t, expectedLocation, dirListHost.String())
	})

	// Test we get dirlist host when PROPFIND returns 404 but dirlisthost set in namespace
	t.Run("test404PropfindValidDirListInNamespace", func(t *testing.T) {
		expectedLocation := "http://origin"
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}
		server := httptest.NewServer(http.HandlerFunc(handler))
		defer server.Close()
		ctx := context.Background()
		testObjectUrl, err := url.Parse("pelican://federation/some/object")
		require.NoError(t, err)

		// Wait until our server is working
		err = server_utils.WaitUntilWorking(ctx, "PROPFIND", server.URL, "testServer", http.StatusNotFound, true)
		require.NoError(t, err)

		dirListHost, err := getDirListHost(ctx, testObjectUrl, namespaces.Namespace{DirListHost: expectedLocation}, server.URL)
		require.NoError(t, err)
		assert.Equal(t, expectedLocation, dirListHost.String())
	})

	// Test we get dirlisthost when we are not using a director and namespace has dirlisthost set
	t.Run("testNoDirectorValidDirListInNamespace", func(t *testing.T) {
		expectedLocation := "http://origin"
		ctx := context.Background()
		testObjectUrl, err := url.Parse("pelican://federation/some/object")
		require.NoError(t, err)
		dirListHost, err := getDirListHost(ctx, testObjectUrl, namespaces.Namespace{DirListHost: expectedLocation}, "")
		require.NoError(t, err)
		assert.Equal(t, expectedLocation, dirListHost.String())
	})

	// Test if PROPFIND and ns.dirlisthost fail, we get dirListingNotSupported error
	t.Run("testInvalidPropfindNoDirListInNamespace", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		server := httptest.NewServer(http.HandlerFunc(handler))
		defer server.Close()
		ctx := context.Background()
		testObjectUrl, err := url.Parse("pelican://federation/some/object")
		require.NoError(t, err)

		// Wait until our server is working
		err = server_utils.WaitUntilWorking(ctx, "PROPFIND", server.URL, "testServer", http.StatusMethodNotAllowed, true)
		require.NoError(t, err)

		_, err = getDirListHost(ctx, testObjectUrl, namespaces.Namespace{}, server.URL)
		require.Error(t, err)
		assert.ErrorIs(t, err, &dirListingNotSupportedError{})
	})

	// Test if PROPFIND if 404 and ns.dirlisthost fail, we get dirListingNotSupported error
	t.Run("test404PropfindNoDirListInNamespace", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}
		server := httptest.NewServer(http.HandlerFunc(handler))
		defer server.Close()
		ctx := context.Background()
		testObjectUrl, err := url.Parse("pelican://federation/some/object")
		require.NoError(t, err)
		_, err = getDirListHost(ctx, testObjectUrl, namespaces.Namespace{}, server.URL)
		require.Error(t, err)
		assert.ErrorIs(t, err, &dirListingNotSupportedError{})
	})

	// Test if no director and namespace doesn't contain dirlisthost, we get dirListingNotSupported error
	t.Run("testNoDirectorNoDirListInNamespace", func(t *testing.T) {
		ctx := context.Background()
		testObjectUrl, err := url.Parse("pelican://federation/some/object")
		require.NoError(t, err)
		_, err = getDirListHost(ctx, testObjectUrl, namespaces.Namespace{}, "")
		require.Error(t, err)
		assert.ErrorIs(t, err, &dirListingNotSupportedError{})
	})

	// Test when director does not return 'location' header (just blank response), we fail
	t.Run("testNoLocationHeaderReturned", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTemporaryRedirect)
		}
		server := httptest.NewServer(http.HandlerFunc(handler))
		defer server.Close()
		ctx := context.Background()
		testObjectUrl, err := url.Parse("pelican://federation/some/object")
		require.NoError(t, err)

		// Wait until our server is working
		err = server_utils.WaitUntilWorking(ctx, "PROPFIND", server.URL, "testServer", http.StatusTemporaryRedirect, true)
		require.NoError(t, err)

		_, err = getDirListHost(ctx, testObjectUrl, namespaces.Namespace{}, server.URL)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "collections URL not found in director response")
	})

	// Test if failure to connect to director we handle that properly
	t.Run("testDirectorFailedToConnect", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			err := json.NewEncoder(w).Encode(map[string]string{"error": "some server error"})
			require.NoError(t, err)
		}
		server := httptest.NewServer(http.HandlerFunc(handler))
		defer server.Close()
		ctx := context.Background()
		testObjectUrl, err := url.Parse("pelican://federation/some/object")
		require.NoError(t, err)

		// Wait until our server is working
		err = server_utils.WaitUntilWorking(ctx, "PROPFIND", server.URL, "testServer", http.StatusServiceUnavailable, true)
		require.NoError(t, err)

		_, err = getDirListHost(ctx, testObjectUrl, namespaces.Namespace{}, server.URL)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "some server error")
	})
}

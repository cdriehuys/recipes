package integration_test

import (
	"bufio"
	"bytes"
	"context"
	"embed"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/cdriehuys/recipes"
	"github.com/cdriehuys/recipes/internal/runtime"
)

// Get the next free unprivileged port. Taken from:
// https://gist.github.com/sevkin/96bdae9274465b2d09191384f86ef39d
func getFreePort() (port int, err error) {
	var a *net.TCPAddr
	if a, err = net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", a); err == nil {
			defer l.Close()
			return l.Addr().(*net.TCPAddr).Port, nil
		}
	}
	return
}

func waitForServer(client *TestClient) error {
	startTime := time.Now()
	ctx := context.Background()

	for {
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"/",
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error making request: %s\n", err.Error())
			continue
		}
		if resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}
		resp.Body.Close()

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if time.Since(startTime) >= 3*time.Second {
				return fmt.Errorf("timeout reached while waiting for endpoint")
			}
			// wait a little while between checks
			time.Sleep(250 * time.Millisecond)
		}
	}
}

var stderrBuff bytes.Buffer

func startServer(t *testing.T, port int) (*TestClient, error) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)

	bindAddr := fmt.Sprintf(":%d", port)
	t.Logf("Starting server on address: %s", bindAddr)

	stderr := bufio.NewWriter(&stderrBuff)
	args := []string{
		"--address", bindAddr,
		"--secret-key", base64.StdEncoding.EncodeToString(make([]byte, 32)),
		"--encryption-key", base64.StdEncoding.EncodeToString(make([]byte, 32)),
	}

	go func() {
		if err := runtime.Run(ctx, stderr, args, embed.FS{}, recipes.TemplateFS); err != nil {
			log.Fatalln(err)
		}
	}()

	// Minimum time required for server to spin up.
	time.Sleep(1 * time.Millisecond)

	client := &TestClient{
		Port:   port,
		Client: &http.Client{},
	}

	if err := waitForServer(client); err != nil {
		return nil, err
	}

	return client, nil
}

func TestSmoke(t *testing.T) {
	port, err := getFreePort()
	if err != nil {
		t.Fatal(err)
	}

	client, err := startServer(t, port)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := client.Get("/")
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200; got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(body))
	t.Log(stderrBuff.String())
}

type TestClient struct {
	Port   int
	Client *http.Client
}

func (c *TestClient) Do(req *http.Request) (*http.Response, error) {
	if req.URL.Host == "" {
		req.URL.Scheme = "http"
		req.URL.Host = c.host()
	}

	return c.Client.Do(req)
}

func (c *TestClient) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	return c.Do(req)
}

func (c *TestClient) host() string {
	return fmt.Sprintf("localhost:%d", c.Port)
}

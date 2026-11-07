package nacos

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/nacos-group/nacos-sdk-go/vo"
)

type mockConfigAPI struct {
	mu          sync.Mutex
	content     string
	getErr      error
	getCalls    int
	listenErr   error
	cancelErr   error
	cancelCalls int
	cancelled   []vo.ConfigParam
	onChange    func(namespace, group, dataID, data string)
	noCallback  bool
}

func (m *mockConfigAPI) GetConfig(param vo.ConfigParam) (string, error) {
	m.mu.Lock()
	m.getCalls++
	m.mu.Unlock()
	if m.getErr != nil {
		return "", m.getErr
	}
	return m.content, nil
}

func (m *mockConfigAPI) ListenConfig(param vo.ConfigParam) error {
	if m.listenErr != nil {
		return m.listenErr
	}
	m.mu.Lock()
	m.onChange = param.OnChange
	m.mu.Unlock()
	if param.OnChange != nil && !m.noCallback {
		param.OnChange("", "", "", "updated")
	}
	return nil
}

func (m *mockConfigAPI) CancelListenConfig(param vo.ConfigParam) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cancelCalls++
	m.cancelled = append(m.cancelled, param)
	return m.cancelErr
}

func (m *mockConfigAPI) cancels() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.cancelCalls
}

func validConfig() Config {
	return Config{
		Address: "127.0.0.1", Port: 8848, Username: "u", Password: "p",
		NamespaceID: "ns", DataID: "d", Group: "g",
	}
}

func TestClient_Download(t *testing.T) {
	c, err := NewClientWithAPI(validConfig(), &mockConfigAPI{content: "yaml"})
	if err != nil {
		t.Fatal(err)
	}
	got, err := c.Download(context.Background())
	if err != nil || got != "yaml" {
		t.Fatalf("Download: %q err=%v", got, err)
	}
}

func TestClient_Download_Error(t *testing.T) {
	c, _ := NewClientWithAPI(validConfig(), &mockConfigAPI{getErr: errors.New("fail")})
	if _, err := c.Download(context.Background()); err == nil {
		t.Fatal("expected error")
	}
}

func TestClient_Watch(t *testing.T) {
	c, _ := NewClientWithAPI(validConfig(), &mockConfigAPI{content: "x"})
	called := false
	if err := c.Watch(context.Background(), func(content []byte) error {
		called = true
		if string(content) != "updated" {
			t.Fatalf("content = %q", content)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("onChange not called")
	}
}

func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("condition not met in time")
}

func TestClient_Watch_CancelledByContext(t *testing.T) {
	api := &mockConfigAPI{noCallback: true}
	c, _ := NewClientWithAPI(validConfig(), api)
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	if err := c.Watch(ctx, func([]byte) error { calls++; return nil }); err != nil {
		t.Fatal(err)
	}
	if api.cancels() != 0 {
		t.Fatal("must not cancel before ctx is done")
	}
	cancel()
	waitFor(t, func() bool { return api.cancels() == 1 })

	api.mu.Lock()
	got := api.cancelled[0]
	cb := api.onChange
	api.mu.Unlock()
	if got.DataId != "d" || got.Group != "g" {
		t.Fatalf("cancel param = %+v", got)
	}
	// Late callbacks after cancellation are dropped.
	cb("", "", "", "late")
	if calls != 0 {
		t.Fatalf("onChange called %d times after cancel", calls)
	}
	// Close afterwards must not cancel twice.
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	if api.cancels() != 1 {
		t.Fatalf("cancels = %d, want 1", api.cancels())
	}
}

func TestClient_Watch_AlreadyCancelledContext(t *testing.T) {
	api := &mockConfigAPI{}
	c, _ := NewClientWithAPI(validConfig(), api)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := c.Watch(ctx, func([]byte) error { return nil }); !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v", err)
	}
}

func TestClient_Watch_Validation(t *testing.T) {
	var nilClient *Client
	if err := nilClient.Watch(context.Background(), func([]byte) error { return nil }); err == nil {
		t.Fatal("nil client")
	}
	c, _ := NewClientWithAPI(validConfig(), &mockConfigAPI{})
	if err := c.Watch(context.Background(), nil); err == nil {
		t.Fatal("nil onChange")
	}
	//nolint:staticcheck // deliberately nil
	if err := c.Watch(nil, func([]byte) error { return nil }); err == nil {
		t.Fatal("nil ctx")
	}
}

func TestClient_Watch_ListenError(t *testing.T) {
	boom := errors.New("listen boom")
	c, _ := NewClientWithAPI(validConfig(), &mockConfigAPI{listenErr: boom})
	if err := c.Watch(context.Background(), func([]byte) error { return nil }); !errors.Is(err, boom) {
		t.Fatalf("err = %v", err)
	}
	// Failed registration must not be tracked.
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestClient_Close_CancelsListeners(t *testing.T) {
	api := &mockConfigAPI{noCallback: true}
	c, _ := NewClientWithAPI(validConfig(), api)
	for i := 0; i < 2; i++ {
		if err := c.Watch(context.Background(), func([]byte) error { return nil }); err != nil {
			t.Fatal(err)
		}
	}
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	if api.cancels() != 2 {
		t.Fatalf("cancels = %d, want 2", api.cancels())
	}
	if err := c.Close(); err != nil || api.cancels() != 2 {
		t.Fatalf("second Close: err=%v cancels=%d", err, api.cancels())
	}
	if err := c.Watch(context.Background(), func([]byte) error { return nil }); !errors.Is(err, ErrClosed) {
		t.Fatalf("Watch after Close err = %v", err)
	}
}

func TestClient_Close_ReturnsCancelError(t *testing.T) {
	boom := errors.New("cancel boom")
	api := &mockConfigAPI{noCallback: true, cancelErr: boom}
	c, _ := NewClientWithAPI(validConfig(), api)
	if err := c.Watch(context.Background(), func([]byte) error { return nil }); err != nil {
		t.Fatal(err)
	}
	if err := c.Close(); !errors.Is(err, boom) {
		t.Fatalf("err = %v", err)
	}
}

func TestClient_Close_NilAndEmpty(t *testing.T) {
	var c *Client
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	c2, _ := NewClientWithAPI(validConfig(), &mockConfigAPI{})
	if err := c2.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestClient_Download_ContextCanceled(t *testing.T) {
	api := &mockConfigAPI{getErr: errors.New("fail")}
	c, _ := NewClientWithAPI(validConfig(), api)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		waitFor(t, func() bool { api.mu.Lock(); defer api.mu.Unlock(); return api.getCalls >= 1 })
		cancel()
	}()
	start := time.Now()
	_, err := c.Download(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v", err)
	}
	if time.Since(start) >= retryDelay {
		t.Fatal("Download did not abort backoff promptly")
	}
}

func TestClient_Download_PreCanceled(t *testing.T) {
	api := &mockConfigAPI{content: "x"}
	c, _ := NewClientWithAPI(validConfig(), api)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := c.Download(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v", err)
	}
	if api.getCalls != 0 {
		t.Fatal("GetConfig must not be called")
	}
}

func TestClient_Download_NotInitialized(t *testing.T) {
	var c *Client
	if _, err := c.Download(context.Background()); err == nil {
		t.Fatal("expected error")
	}
}

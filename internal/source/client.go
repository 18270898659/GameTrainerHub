package source

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"GameTrainerHub/internal/logger"
)

type HttpClient struct {
	client       *http.Client
	rateLimitMs  int
	lastReqTime  time.Time
	mu           sync.Mutex
	userAgent    string
}

func NewHttpClient(proxyURLStr string, rateLimitMs int) (*HttpClient, error) {
	if rateLimitMs <= 0 {
		rateLimitMs = 500
	}

	transport := &http.Transport{
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		DisableCompression: false,
	}

	if proxyURLStr != "" {
		proxyURL, err := url.Parse(proxyURLStr)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %w", err)
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	return &HttpClient{
		client: &http.Client{
			Transport: transport,
			Timeout:   15 * time.Second,
		},
		rateLimitMs: rateLimitMs,
		userAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	}, nil
}

// Get 执行带有速率限制和自动重试的 GET 请求
func (c *HttpClient) Get(ctx context.Context, targetURL string, referer string) ([]byte, error) {
	var body []byte
	var err error

	for attempt := 1; attempt <= 3; attempt++ {
		c.throttle()

		req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
		if reqErr != nil {
			return nil, reqErr
		}

		req.Header.Set("User-Agent", c.userAgent)
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		req.Header.Set("Accept-Language", "en-US,en;q=0.9,zh-CN;q=0.8")
		if referer != "" {
			req.Header.Set("Referer", referer)
		}

		resp, doErr := c.client.Do(req)
		if doErr != nil {
			err = doErr
			logger.Warn("HTTP request failed, retrying", "url", targetURL, "attempt", attempt, "err", doErr)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(attempt*500) * time.Millisecond):
				continue
			}
		}

		if resp.StatusCode >= 500 || resp.StatusCode == 429 {
			resp.Body.Close()
			err = fmt.Errorf("HTTP status %d", resp.StatusCode)
			logger.Warn("Server busy or error, retrying", "url", targetURL, "status", resp.StatusCode, "attempt", attempt)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(attempt*1000) * time.Millisecond):
				continue
			}
		}

		body, err = io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("HTTP error status: %d", resp.StatusCode)
		}

		return body, nil
	}

	return nil, fmt.Errorf("request failed after 3 attempts: %w", err)
}

// Head 执行 HEAD 请求（常用于探测 Range 支持与获取文件元信息）
func (c *HttpClient) Head(ctx context.Context, targetURL string, referer string) (*http.Response, error) {
	c.throttle()

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, targetURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", c.userAgent)
	if referer != "" {
		req.Header.Set("Referer", referer)
	}

	return c.client.Do(req)
}

// RawClient 获取底层 *http.Client
func (c *HttpClient) RawClient() *http.Client {
	return c.client
}

func (c *HttpClient) throttle() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(c.lastReqTime)
	waitDuration := time.Duration(c.rateLimitMs)*time.Millisecond - elapsed
	if waitDuration > 0 {
		time.Sleep(waitDuration)
	}
	c.lastReqTime = time.Now()
}


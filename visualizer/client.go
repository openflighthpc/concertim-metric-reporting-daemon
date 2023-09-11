package visualizer

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/alces-flight/concertim-metric-reporting-daemon/config"
	"github.com/go-chi/jwtauth/v5"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

var AuthenticationError = fmt.Errorf("authentication failed")

// Client provides methods for interacting with the Concertim Visualizer
// API.
type Client struct {
	authToken string
	client    *http.Client
	config    config.VisualizerAPI
	logger    zerolog.Logger
	tokenAuth *jwtauth.JWTAuth
}

func New(logger zerolog.Logger, config config.VisualizerAPI) *Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: config.SkipCertificateCheck},
	}
	client := &http.Client{Transport: tr}
	return &Client{
		config:    config,
		client:    client,
		logger:    logger.With().Str("component", "visualizerAPI").Logger(),
		tokenAuth: jwtauth.New("HS256", config.JWTSecret, nil),
	}
}

func (v *Client) authenticate() error {
	if v.authToken != "" {
		_, err := jwtauth.VerifyToken(v.tokenAuth, v.authToken)
		if err == nil {
			v.logger.Debug().Msg("using existing auth token")
			return nil
		}
		v.logger.Debug().Err(err).Msg("existing token invalid")
		v.authToken = ""
	}
	v.logger.Debug().Str("url", v.config.AuthUrl).Msg("authenticating")
	credentials := map[string]map[string]string{
		"user": {"login": v.config.Username, "password": v.config.Password},
	}
	body, _ := json.Marshal(credentials)
	resp, err := v.Do("POST", v.config.AuthUrl, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("%w: %s", AuthenticationError, err.Error())
	}
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return fmt.Errorf("%w: %s: %s", AuthenticationError, v.config.AuthUrl, resp.Status)
	}
	headerName := "Authorization"
	v.authToken = strings.Fields(resp.Header.Get(headerName))[1]
	return nil
}

func (v *Client) GetDSM() ([]byte, error) {
	url := v.config.DataSourceMapUrl
	v.logger.Debug().Str("url", url).Msg("getting dsms")
	resp, err := v.Get(url)
	if err != nil {
		return nil, augmentError(err, "visualizerAPI.GetDSM", "GET", url)
	}
	if resp.StatusCode != 200 {
		msg := fmt.Sprintf("visualizerAPI.GetDSM failed: GET %s: %s", url, resp.Status)
		return nil, errors.New(msg)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, augmentError(err, "reading response", "GET", url)
	}
	return data, nil
}

func (v *Client) Get(url string) (*http.Response, error) {
	err := v.authenticate()
	if err != nil {
		return nil, augmentError(err, "", "GET", url)
	}
	return v.Do("GET", url, nil)
}

func (v *Client) Post(url string, body io.Reader) (*http.Response, error) {
	err := v.authenticate()
	if err != nil {
		return nil, augmentError(err, "", "POST", url)
	}
	return v.Do("POST", url, body)
}

func (v *Client) Do(method, url string, body io.Reader) (*http.Response, error) {
	v.logger.Debug().Str("method", method).Str("url", url).Msg("sending request")
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, errors.Wrap(err, "unexpected")
	}
	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}
	req.Header.Add("Accept", "application/json")
	if v.authToken != "" {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", v.authToken))
	}
	resp, err := v.client.Do(req)
	if err != nil {
		return nil, augmentError(err, "", "GET", url)
	}
	return resp, nil
}

func augmentError(err error, msg, method, url string) error {
	if !errors.Is(err, AuthenticationError) && !strings.Contains(err.Error(), url) {
		if msg == "" {
			msg = fmt.Sprintf("%s %s", method, url)
		} else {
			msg = fmt.Sprintf("%s: %s %s", msg, method, url)
		}
	}
	if msg == "" {
		return err
	} else {
		return errors.Wrap(err, msg)
	}
}

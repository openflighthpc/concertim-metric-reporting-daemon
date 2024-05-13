//==============================================================================
// Copyright (C) 2024-present Alces Flight Ltd.
//
// This file is part of Concertim Metric Reporting Daemon.
//
// This program and the accompanying materials are made available under
// the terms of the Eclipse Public License 2.0 which is available at
// <https://www.eclipse.org/legal/epl-2.0>, or alternative license
// terms made available by Alces Flight Ltd - please direct inquiries
// about licensing to licensing@alces-flight.com.
//
// Concertim Metric Reporting Daemon is distributed in the hope that it will be useful, but
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, EITHER EXPRESS OR
// IMPLIED INCLUDING, WITHOUT LIMITATION, ANY WARRANTIES OR CONDITIONS
// OF TITLE, NON-INFRINGEMENT, MERCHANTABILITY OR FITNESS FOR A
// PARTICULAR PURPOSE. See the Eclipse Public License 2.0 for more
// details.
//
// You should have received a copy of the Eclipse Public License 2.0
// along with Concertim Metric Reporting Daemon. If not, see:
//
//  https://opensource.org/licenses/EPL-2.0
//
// For more information on Concertim Metric Reporting Daemon, please visit:
// https://github.com/openflighthpc/concertim-metric-reporting-daemon
//==============================================================================

package visualizer

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/jwtauth/v5"
	"github.com/openflighthpc/concertim-metric-reporting-daemon/config"
	"github.com/openflighthpc/concertim-metric-reporting-daemon/domain"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

var AuthenticationError = fmt.Errorf("authentication failed")

// Client provides methods for interacting with the Concertim Visualizer
// API.
type Client struct {
	authToken string
	client    *http.Client
	Config    config.VisualizerAPI
	logger    zerolog.Logger
	tokenAuth *jwtauth.JWTAuth
}

func New(logger zerolog.Logger, config config.VisualizerAPI) *Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: config.SkipCertificateCheck},
	}
	client := &http.Client{Transport: tr}
	return &Client{
		Config:    config,
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
	v.logger.Debug().Str("url", v.Config.AuthUrl).Msg("authenticating")
	credentials := map[string]map[string]string{
		"user": {"login": v.Config.Username, "password": v.Config.Password},
	}
	body, _ := json.Marshal(credentials)
	resp, err := v.Do("POST", v.Config.AuthUrl, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("%w: %s", AuthenticationError, err.Error())
	}
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return fmt.Errorf("%w: %s: %s", AuthenticationError, v.Config.AuthUrl, resp.Status)
	}
	headerName := "Authorization"
	v.authToken = strings.Fields(resp.Header.Get(headerName))[1]
	return nil
}

func (v *Client) GetDSM() (map[domain.HostId]domain.DSM, map[domain.DSM]domain.HostId, error) {
	url := v.Config.DataSourceMapUrl
	v.logger.Debug().Str("url", url).Msg("getting dsms")
	resp, err := v.Get(url)
	if err != nil {
		return nil, nil, augmentError(err, "visualizerAPI.GetDSM", "GET", url)
	}
	if resp.StatusCode != 200 {
		msg := fmt.Sprintf("visualizerAPI.GetDSM failed: GET %s: %s", url, resp.Status)
		return nil, nil, errors.New(msg)
	}
	defer func() { _ = resp.Body.Close() }()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, augmentError(err, "reading response", "GET", url)
	}
	parser := Parser{Logger: v.logger}
	hostIdToDSM, dsmToHostId, err := parser.ParseDSM(data)
	if err != nil {
		return nil, nil, errors.Wrap(err, "parsing DSM")
	}
	return hostIdToDSM, dsmToHostId, nil
}

func (v *Client) Get(url string) (*http.Response, error) {
	doAuthenticatedRequest := func() (*http.Response, error) {
		err := v.authenticate()
		if err != nil {
			return nil, augmentError(err, "", "GET", url)
		}
		return v.Do("GET", url, nil)
	}
	resp, err := doAuthenticatedRequest()
	if err != nil {
		return resp, err
	} else if resp.StatusCode == 401 {
		v.logger.Debug().Msg("existing auth token appears invalid")
		v.authToken = ""
		return doAuthenticatedRequest()
	}
	return resp, nil
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
		return resp, augmentError(err, "", method, url)
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

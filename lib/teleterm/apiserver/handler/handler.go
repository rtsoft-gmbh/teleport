// Copyright 2021 Gravitational, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package handler

import (
	"context"

	api "github.com/gravitational/teleport/lib/teleterm/api/protogen/golang/v1"
	"github.com/gravitational/teleport/lib/teleterm/daemon"
)

// Config is the terminal service
type Config struct {
	// DaemonService is the instance of daemon service
	DaemonService *daemon.Service
}

// Handler implements teleterm api service
type Handler struct {
	// Config is the service config
	Config
}

func New(cfg Config) (*Handler, error) {
	return &Handler{
		cfg,
	}, nil
}

func (s *Handler) CreateClusterLoginChallenge(context.Context, *api.CreateClusterLoginChallengeRequest) (*api.ClusterLoginChallenge, error) {
	return nil, nil
}

func (s *Handler) SolveClusterLoginChallenge(context.Context, *api.SolveClusterLoginChallengeRequest) (*api.SolveClusterLoginChallengeResponse, error) {
	return nil, nil
}

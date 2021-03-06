// Copyright 2017 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package na

import (
	"fmt"

	"cloud.google.com/go/compute/metadata"
	"github.com/golang/glog"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	cred "istio.io/auth/pkg/credential"
)

const (
	bearerTokenScheme = "Bearer"
	httpAuthHeader    = "authorization"
)

type jwtAccess struct {
	token string
}

func (j *jwtAccess) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		httpAuthHeader: fmt.Sprintf("%s %s", bearerTokenScheme, j.token),
	}, nil
}

func (j *jwtAccess) RequireTransportSecurity() bool {
	return true
}

type gcpPlatformImpl struct {
	fetcher cred.TokenFetcher
}

func (na *gcpPlatformImpl) IsProperPlatform() bool {
	return metadata.OnGCE()
}

func (na *gcpPlatformImpl) GetDialOptions(cfg *Config) ([]grpc.DialOption, error) {
	jwtKey, err := na.fetcher.FetchToken()
	if err != nil {
		glog.Errorf("Failed to get instance from GCE metadata %s, please make sure this binary is running on a GCE VM", err)
		return nil, err
	}

	creds, err := credentials.NewClientTLSFromFile(cfg.RootCACertFile, "")
	if err != nil {
		return nil, err
	}

	options := []grpc.DialOption{grpc.WithPerRPCCredentials(&jwtAccess{jwtKey}), grpc.WithTransportCredentials(creds)}
	return options, nil
}

func (na *gcpPlatformImpl) GetServiceIdentity() (string, error) {
	// TODO(wattli): update this once we are ready for GCE
	return "", nil
}

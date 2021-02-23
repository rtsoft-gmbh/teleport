/*
Copyright 2018-2021 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"

	"github.com/gravitational/teleport/api/client"
	"github.com/gravitational/teleport/api/types"
	"github.com/gravitational/teleport/lib/auth"

	"github.com/pborman/uuid"
)

func main() {
	ctx := context.Background()
	log.Printf("Starting Teleport client...")

	tlsConfig := &tls.Config{}
	// Create valid tlsConfig here to load TLS credentials.

	clt, err := auth.NewClient(client.Config{
		// Addrs is the Auth Server address, must be local.
		Addrs: []string{"localhost:3025"},
		// Multiple credentials can be provided, and the first credentials to
		// successfully authenticate an open connection to the client will be used.
		Credentials: []client.Credentials{
			client.LoadKeyPair(
				"certs/access-admin.crt",
				"certs/access-admin.key",
				"certs/access-admin.cas",
			),
			client.LoadIdentityFile("certs/access-admin-identity"),
			client.LoadTLS(tlsConfig),
		},
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer clt.Close()

	if err := demoClient(ctx, clt); err != nil {
		log.Printf("error(s) in demoClient: %v", err)
	}
}

func demoClient(ctx context.Context, clt *auth.Client) (err error) {
	// Create a new access request for the `access-admin` user to use the `admin` role.
	accessReq, err := types.NewAccessRequest(uuid.New(), "access-admin", "admin")
	if err != nil {
		return fmt.Errorf("failed to make new access request: %v", err)
	}
	if err = clt.CreateAccessRequest(ctx, accessReq); err != nil {
		return fmt.Errorf("failed to create access request: %v", err)
	}
	log.Printf("Created access request: %v", accessReq)

	defer func() {
		if err2 := clt.DeleteAccessRequest(ctx, accessReq.GetName()); err2 != nil {
			err = fmt.Errorf("%v\nfailed to delete access request: %v", err, err2)
			return
		}
		log.Println("Deleted access request")
	}()

	// Approve the access request as if this was another party.
	if err = clt.SetAccessRequestState(ctx, types.AccessRequestUpdate{
		RequestID: accessReq.GetName(),
		State:     types.RequestState_APPROVED,
	}); err != nil {
		return fmt.Errorf("failed to accept request: %v", err)
	}
	log.Printf("Approved access request")

	return nil
}

/**
 * Copyright (C) 2015 Red Hat, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *         http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
// Copyright 2017 The etcd Authors
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

package integration

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"golang.org/x/net/context"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/etcdserver/api/v3rpc/rpctypes"
	pb "github.com/coreos/etcd/etcdserver/etcdserverpb"
	"github.com/coreos/etcd/pkg/testutil"
)

// TestV3AuthEmptyUserGet ensures that a get with an empty user will return an empty user error.
func TestV3AuthEmptyUserGet(t *testing.T) {
	defer testutil.AfterTest(t)
	clus := NewClusterV3(t, &ClusterConfig{Size: 1})
	defer clus.Terminate(t)

	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()

	api := toGRPC(clus.Client(0))
	authSetupRoot(t, api.Auth)

	_, err := api.KV.Range(ctx, &pb.RangeRequest{Key: []byte("abc")})
	if !eqErrGRPC(err, rpctypes.ErrUserEmpty) {
		t.Fatalf("got %v, expected %v", err, rpctypes.ErrUserEmpty)
	}
}

// TestV3AuthTokenWithDisable tests that auth won't crash if
// given a valid token when authentication is disabled
func TestV3AuthTokenWithDisable(t *testing.T) {
	defer testutil.AfterTest(t)
	clus := NewClusterV3(t, &ClusterConfig{Size: 1})
	defer clus.Terminate(t)

	authSetupRoot(t, toGRPC(clus.Client(0)).Auth)

	c, cerr := clientv3.New(clientv3.Config{Endpoints: clus.Client(0).Endpoints(), Username: "root", Password: "123"})
	if cerr != nil {
		t.Fatal(cerr)
	}
	defer c.Close()

	rctx, cancel := context.WithCancel(context.TODO())
	donec := make(chan struct{})
	go func() {
		defer close(donec)
		for rctx.Err() == nil {
			c.Put(rctx, "abc", "def")
		}
	}()

	time.Sleep(10 * time.Millisecond)
	if _, err := c.AuthDisable(context.TODO()); err != nil {
		t.Fatal(err)
	}
	time.Sleep(10 * time.Millisecond)

	cancel()
	<-donec
}

func TestV3AuthRevision(t *testing.T) {
	defer testutil.AfterTest(t)
	clus := NewClusterV3(t, &ClusterConfig{Size: 1})
	defer clus.Terminate(t)

	api := toGRPC(clus.Client(0))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	presp, perr := api.KV.Put(ctx, &pb.PutRequest{Key: []byte("foo"), Value: []byte("bar")})
	cancel()
	if perr != nil {
		t.Fatal(perr)
	}
	rev := presp.Header.Revision

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	aresp, aerr := api.Auth.UserAdd(ctx, &pb.AuthUserAddRequest{Name: "root", Password: "123"})
	cancel()
	if aerr != nil {
		t.Fatal(aerr)
	}
	if aresp.Header.Revision != rev {
		t.Fatalf("revision expected %d, got %d", rev, aresp.Header.Revision)
	}
}

func authSetupRoot(t *testing.T, auth pb.AuthClient) {
	if _, err := auth.UserAdd(context.TODO(), &pb.AuthUserAddRequest{Name: "root", Password: "123"}); err != nil {
		t.Fatal(err)
	}
	if _, err := auth.RoleAdd(context.TODO(), &pb.AuthRoleAddRequest{Name: "root"}); err != nil {
		t.Fatal(err)
	}
	if _, err := auth.UserGrantRole(context.TODO(), &pb.AuthUserGrantRoleRequest{User: "root", Role: "root"}); err != nil {
		t.Fatal(err)
	}
	if _, err := auth.AuthEnable(context.TODO(), &pb.AuthEnableRequest{}); err != nil {
		t.Fatal(err)
	}
}

func TestV3AuthOldRevConcurrent(t *testing.T) {
	defer testutil.AfterTest(t)
	clus := NewClusterV3(t, &ClusterConfig{Size: 1})
	defer clus.Terminate(t)

	authSetupRoot(t, toGRPC(clus.Client(0)).Auth)

	c, cerr := clientv3.New(clientv3.Config{
		Endpoints:   clus.Client(0).Endpoints(),
		DialTimeout: 5 * time.Second,
		Username:    "root",
		Password:    "123",
	})
	testutil.AssertNil(t, cerr)
	defer c.Close()

	var wg sync.WaitGroup
	f := func(i int) {
		defer wg.Done()
		role, user := fmt.Sprintf("test-role-%d", i), fmt.Sprintf("test-user-%d", i)
		_, err := c.RoleAdd(context.TODO(), role)
		testutil.AssertNil(t, err)
		_, err = c.RoleGrantPermission(context.TODO(), role, "", clientv3.GetPrefixRangeEnd(""), clientv3.PermissionType(clientv3.PermReadWrite))
		testutil.AssertNil(t, err)
		_, err = c.UserAdd(context.TODO(), user, "123")
		testutil.AssertNil(t, err)
		_, err = c.Put(context.TODO(), "a", "b")
		testutil.AssertNil(t, err)
	}
	// needs concurrency to trigger
	numRoles := 2
	wg.Add(numRoles)
	for i := 0; i < numRoles; i++ {
		go f(i)
	}
	wg.Wait()
}

// TestV3AuthWithLeaseRevokeWithRoot ensures that granted leases
// with root user be revoked after TTL.
func TestV3AuthWithLeaseRevokeWithRoot(t *testing.T) {
	defer testutil.AfterTest(t)

	clus := NewClusterV3(t, &ClusterConfig{Size: 1})
	defer clus.Terminate(t)

	api := toGRPC(clus.Client(0))
	authSetupRoot(t, api.Auth)

	rootc, cerr := clientv3.New(clientv3.Config{
		Endpoints: clus.Client(0).Endpoints(),
		Username:  "root",
		Password:  "123",
	})
	if cerr != nil {
		t.Fatal(cerr)
	}
	defer rootc.Close()

	leaseResp, err := rootc.Grant(context.TODO(), 2)
	if err != nil {
		t.Fatal(err)
	}
	leaseID := leaseResp.ID

	if _, err = rootc.Put(context.TODO(), "foo", "bar", clientv3.WithLease(leaseID)); err != nil {
		t.Fatal(err)
	}

	// wait for lease expire
	time.Sleep(3 * time.Second)

	tresp, terr := api.Lease.LeaseTimeToLive(
		context.TODO(),
		&pb.LeaseTimeToLiveRequest{
			ID:   int64(leaseID),
			Keys: true,
		},
	)
	if terr != nil {
		t.Error(terr)
	}
	if len(tresp.Keys) > 0 || tresp.GrantedTTL != 0 {
		t.Errorf("lease %016x should have been revoked, got %+v", leaseID, tresp)
	}
	if tresp.TTL != -1 {
		t.Errorf("lease %016x should have been expired, got %+v", leaseID, tresp)
	}
}

//--------------------------------------------------------------------------
// Copyright 2018 Infinite Devices GmbH
// www.infinimesh.io
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.
//--------------------------------------------------------------------------

package dgraph

import (
	"context"
	"testing"

	"os"

	randomdata "github.com/Pallinder/go-randomdata"
	"github.com/dgraph-io/dgo"
	"github.com/dgraph-io/dgo/protos/api"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/infinimesh/infinimesh/pkg/node"
	"github.com/infinimesh/infinimesh/pkg/node/nodepb"
)

var repo node.Repo

func init() {
	dgURL := os.Getenv("DGRAPH_URL")
	if dgURL == "" {
		dgURL = "localhost:9080"
	}
	conn, err := grpc.Dial(dgURL, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	dg := dgo.NewDgraphClient(api.NewDgraphClient(conn))
	repo = NewDGraphRepo(dg)
}

func TestAuthorize(t *testing.T) {
	ctx := context.Background()
	_, err := repo.CreateNamespace(ctx, "default")
	require.NoError(t, err)

	account, err := repo.CreateUserAccount(ctx, randomdata.SillyName(), "password", false, true)
	require.NoError(t, err)

	node, err := repo.CreateObject(ctx, "sample-node", "", "asset", "default")
	require.NoError(t, err)

	err = repo.Authorize(ctx, account, node, "READ", true)
	require.NoError(t, err)

	decision, err := repo.IsAuthorized(ctx, node, account, "READ")
	require.NoError(t, err)
	require.True(t, decision)
}

func TestIsAuthorizedNamespace(t *testing.T) {
	ctx := context.Background()

	accountname := randomdata.SillyName()
	account, err := repo.CreateUserAccount(ctx, accountname, "password", false, true)
	require.NoError(t, err)

	ns, err := repo.GetNamespace(ctx, accountname)
	require.NoError(t, err)

	decision, err := repo.IsAuthorizedNamespace(ctx, ns.Id, account, nodepb.Action_WRITE)
	require.NoError(t, err)
	require.True(t, decision)
}

func TestListInNamespaceForAccount(t *testing.T) {
	ctx := context.Background()

	acc := randomdata.SillyName()

	// Create Account
	account, err := repo.CreateUserAccount(ctx, acc, "password", false, true)
	require.NoError(t, err)

	//Get Namespace
	nsName, err := repo.GetNamespace(ctx, acc)

	//Create Object
	newObj, err := repo.CreateObject(ctx, "sample-node", "", "asset", nsName.Name)
	require.NoError(t, err)

	err = repo.AuthorizeNamespace(ctx, account, nsName.Id, nodepb.Action_WRITE)
	require.NoError(t, err)

	objs, err := repo.ListForAccount(ctx, account, nsName.GetName(), true)
	require.NoError(t, err)

	// Assert
	require.Contains(t, objs, &nodepb.Object{Uid: newObj, Name: "sample-node", Kind: "asset", Objects: []*nodepb.Object{}})
}

func TestChangePassword(t *testing.T) {
	ctx := context.Background()

	acc := randomdata.SillyName()

	// Create Account
	_, err := repo.CreateUserAccount(ctx, acc, "password", false, true)
	require.NoError(t, err)

	err = repo.SetPassword(ctx, acc, "newpassword")
	require.NoError(t, err)

	ok, _, _, err := repo.Authenticate(ctx, acc, "newpassword")
	require.True(t, ok)
}

func TestChangePasswordWithNoUser(t *testing.T) {
	ctx := context.Background()

	err := repo.SetPassword(ctx, "non-existing-user", "newpassword")
	require.Error(t, err)
}

func TestListPermissionsOnNamespace(t *testing.T) {
	ctx := context.Background()

	randomUser := randomdata.SillyName()

	//Create Account
	accountID, err := repo.CreateUserAccount(ctx, randomUser, "password", false, true)
	require.NoError(t, err)

	//Get Namespace
	nsName, err := repo.GetNamespace(ctx, randomUser)

	err = repo.AuthorizeNamespace(ctx, accountID, nsName.Id, nodepb.Action_WRITE)
	require.NoError(t, err)

	permissions, err := repo.ListPermissionsInNamespace(ctx, nsName.Id)
	require.NoError(t, err)

	var namespaceFound bool
	for _, permission := range permissions {
		if permission.AccountName == nsName.Name {
			namespaceFound = true
		}
	}
	require.True(t, namespaceFound)
}

func TestDeletePermissionOnNamespace(t *testing.T) {
	ctx := context.Background()

	randomUser := randomdata.SillyName()

	//Create Account
	accountID, err := repo.CreateUserAccount(ctx, randomUser, "password", false, true)
	require.NoError(t, err)

	//Get Namespace
	nsName, err := repo.GetNamespace(ctx, randomUser)

	err = repo.AuthorizeNamespace(ctx, accountID, nsName.Id, nodepb.Action_WRITE)
	require.NoError(t, err)

	err = repo.DeletePermissionInNamespace(ctx, nsName.Id, accountID)
	require.NoError(t, err)

	permissions, err := repo.ListPermissionsInNamespace(ctx, nsName.Id)
	require.NoError(t, err)
	require.Empty(t, permissions)

}

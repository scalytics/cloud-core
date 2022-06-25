/*
Copyright © 2021-2022 Infinite Devices GmbH, Nikita Ivanovski info@slnt-opp.xyz

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
package graph

import (
	"context"

	"github.com/arangodb/go-driver"
	"go.uber.org/zap"

	"github.com/infinimesh/infinimesh/pkg/graph/schema"
	"github.com/infinimesh/proto/node/access"
	pb "github.com/infinimesh/proto/plugins"
)

type Plugin struct {
	*pb.Plugin
	driver.DocumentMeta
}

func (o *Plugin) ID() driver.DocumentID {
	return o.DocumentMeta.ID
}

func (o *Plugin) SetAccessLevel(level access.Level) {
	if o.Access == nil {
		o.Access = &access.Access{
			Level: level,
		}
		return
	}
	o.Access.Level = level
}

func NewBlankPluginDocument(key string) *Plugin {
	return &Plugin{
		Plugin: &pb.Plugin{
			Uuid: key,
		},
		DocumentMeta: NewBlankDocument(schema.PLUGINS_COL, key),
	}
}

type PluginsController struct {
	pb.UnimplementedPluginsServiceServer
	log *zap.Logger

	col driver.Collection // Plugins Collection

	db driver.Database
}

func NewPluginsController(log *zap.Logger, db driver.Database) *PluginsController {
	ctx := context.TODO()
	col, _ := db.Collection(ctx, schema.PLUGINS_COL)
	return &PluginsController{
		log: log.Named("PluginsController"), col: col, db: db,
	}
}

func ValidateRoot(ctx context.Context) bool {
	rootV := ctx.Value(inf.InfinimeshRootCtxKey)
	if rootV == nil {
		return false
	}

	root, ok := rootV.(bool)
	return ok && root
}

func ValidatePluginDocument(p *pb.Plugin) string {
	if p.Title == "" {
		return "Title cannot be empty"
	}

	if p.Kind == pb.PluginKind_UNKNOWN {
		return "Kind can't be Unknown"
	} else if p.Kind == pb.PluginKind_EMBEDDED && p.EmbeddedConf == nil {
		return "Kind is set to Embedded, but no conf provided"
	}

	return ""
}

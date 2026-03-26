/*
 * Copyright 2025 coze-dev Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package impl

import (
	"context"
	"fmt"
	"runtime"

	"github.com/coze-dev/coze-studio/backend/api/model/admin/config"
	"github.com/coze-dev/coze-studio/backend/infra/document/searchstore"
	"github.com/coze-dev/coze-studio/backend/infra/document/searchstore/impl/elasticsearch"
	"github.com/coze-dev/coze-studio/backend/infra/es/impl/es"
)

type Manager = searchstore.Manager

func New(ctx context.Context, conf *config.KnowledgeConfig, es es.Client) ([]Manager, error) {
	// es full text search
	esSearchstoreManager := elasticsearch.NewManager(&elasticsearch.ManagerConfig{Client: es})

	// On Windows, only return Elasticsearch manager to avoid Milvus dependencies
	if runtime.GOOS == "windows" {
		return []searchstore.Manager{esSearchstoreManager}, nil
	}

	// vector search
	mgr, err := getVectorStore(ctx, conf)
	if err != nil {
		return nil, fmt.Errorf("init vector store failed, err=%w", err)
	}

	return []searchstore.Manager{esSearchstoreManager, mgr}, nil
}

func getVectorStore(ctx context.Context, conf *config.KnowledgeConfig) (searchstore.Manager, error) {
	return nil, fmt.Errorf("vector store not supported on this platform")
}

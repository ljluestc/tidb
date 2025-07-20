// Copyright 2016 PingCAP, Inc.
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

package executor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pingcap/tidb/pkg/expression"
	"github.com/pingcap/tidb/pkg/kv"
	"github.com/pingcap/tidb/pkg/model"
	"github.com/pingcap/tidb/pkg/parser/mysql"
	"github.com/pingcap/tidb/pkg/sessionctx"
	"github.com/pingcap/tidb/pkg/terror"
	"github.com/pingcap/tidb/pkg/types"
	"github.com/pingcap/tidb/pkg/util"
	"github.com/pingcap/tidb/pkg/util/chunk"
	"github.com/pingcap/tidb/pkg/util/memory"
)

// Executor executes a query.
type Executor interface {
	Next(ctx context.Context, req *chunk.Chunk) error
	Close() error
	Open(context.Context) error
	Schema() *expression.Schema
}

// ShowResultInPlugin is used to show information in plugin.
// The main usage of this interface is to expose the information to other component.
type ShowResultInPlugin interface {
	// GetPluginResult returns a list of data for plugin to process.
	GetPluginResult() [][]types.Datum
}

// GetLackHandles gets the handles lack in the result.
func GetLackHandles(expectedHandles []kv.Handle, handlesMap kv.HandleMap) []kv.Handle {
	if handlesMap.Len() == 0 {
		return expectedHandles
	}

	var lackHandles []kv.Handle
	for _, h := range expectedHandles {
		var ok bool
		h, ok = handlesMap.Get(h)
		if !ok {
			lackHandles = append(lackHandles, h)
		} else {
			handlesMap.Delete(h)
		}
	}
	return lackHandles
}

// IndexLookUpRunTimeStats record the run time stats for index lookup executor.
type IndexLookUpRunTimeStats struct {
	FetchHandleTotal         int64
	FetchHandle              int64
	BuildHashTable           int64
	TaskWait                 int64
	TableRowScan             int64
	TableTaskNum             int
	Concurrency              int
	NextWaitIndexScan        time.Duration
	NextWaitTableLookUpBuild time.Duration
	NextWaitTableLookUpResp  time.Duration
}

// Clone implements the RuntimeStats interface.
func (i *IndexLookUpRunTimeStats) Clone() *IndexLookUpRunTimeStats {
	newIndexLookUpRunTimeStats := *i
	return &newIndexLookUpRunTimeStats
}

// String implements the RuntimeStats interface.
func (i *IndexLookUpRunTimeStats) String() string {
	return fmt.Sprintf("index_task: {total_time: %s, fetch_handle: %s, build: %s, wait: %s}"+
		", table_task: {total_time: %s, num: %d, concurrency: %d}"+
		", next: {wait_index: %s, wait_table_lookup_build: %s, wait_table_lookup_resp: %s}",
		util.FormatDuration(time.Duration(i.FetchHandleTotal)),
		util.FormatDuration(time.Duration(i.FetchHandle)),
		util.FormatDuration(time.Duration(i.BuildHashTable)),
		util.FormatDuration(time.Duration(i.TaskWait)),
		util.FormatDuration(time.Duration(i.TableRowScan)),
		i.TableTaskNum, i.Concurrency,
		i.NextWaitIndexScan.String(),
		i.NextWaitTableLookUpBuild.String(),
		i.NextWaitTableLookUpResp.String(),
	)
}

// Merge implements the RuntimeStats interface.
func (i *IndexLookUpRunTimeStats) Merge(other *IndexLookUpRunTimeStats) {
	i.FetchHandleTotal += other.FetchHandleTotal
	i.FetchHandle += other.FetchHandle
	i.BuildHashTable += other.BuildHashTable
	i.TaskWait += other.TaskWait
	i.TableRowScan += other.TableRowScan
	i.TableTaskNum += other.TableTaskNum
	i.NextWaitIndexScan += other.NextWaitIndexScan
	i.NextWaitTableLookUpBuild += other.NextWaitTableLookUpBuild
	i.NextWaitTableLookUpResp += other.NextWaitTableLookUpResp
}

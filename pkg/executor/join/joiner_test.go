// Copyright 2019 PingCAP, Inc.
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

package join

import (
	"math/rand"
	"sync"
	"testing"

	"github.com/pingcap/tidb/pkg/domain"
	"github.com/pingcap/tidb/pkg/parser/mysql"
	"github.com/pingcap/tidb/pkg/planner/core/base"
	"github.com/pingcap/tidb/pkg/sessionctx"
	"github.com/pingcap/tidb/pkg/sessionctx/vardef"
	"github.com/pingcap/tidb/pkg/types"
	"github.com/pingcap/tidb/pkg/util/chunk"
	"github.com/pingcap/tidb/pkg/util/disk"
	"github.com/pingcap/tidb/pkg/util/memory"
	"github.com/pingcap/tidb/pkg/util/mvmap"
	"github.com/pingcap/tidb/pkg/util/mock"
	"github.com/stretchr/testify/require"
)

func defaultCtx() sessionctx.Context {
	ctx := mock.NewContext()
	ctx.GetSessionVars().InitChunkSize = vardef.DefInitChunkSize
	ctx.GetSessionVars().MaxChunkSize = vardef.DefMaxChunkSize
	ctx.GetSessionVars().StmtCtx.MemTracker = memory.NewTracker(-1, ctx.GetSessionVars().MemQuotaQuery)
	ctx.GetSessionVars().StmtCtx.DiskTracker = disk.NewTracker(-1, -1)
	ctx.GetSessionVars().SnapshotTS = uint64(1)
	ctx.BindDomainAndSchValidator(domain.NewMockDomain(), nil)
	return ctx
}

func TestRequiredRows(t *testing.T) {
	joinTypes := []base.JoinType{base.InnerJoin, base.LeftOuterJoin, base.RightOuterJoin}
	lTypes := [][]byte{
		{mysql.TypeLong},
		{mysql.TypeFloat},
		{mysql.TypeLong, mysql.TypeFloat},
	}
	rTypes := lTypes

	convertTypes := func(mysqlTypes []byte) []*types.FieldType {
		fieldTypes := make([]*types.FieldType, 0, len(mysqlTypes))
		for _, t := range mysqlTypes {
			fieldTypes = append(fieldTypes, types.NewFieldType(t))
		}
		return fieldTypes
	}

	for _, joinType := range joinTypes {
		for _, ltype := range lTypes {
			for _, rtype := range rTypes {
				maxChunkSize := defaultCtx().GetSessionVars().MaxChunkSize
				lfields := convertTypes(ltype)
				rfields := convertTypes(rtype)
				outerRow := genTestChunk(maxChunkSize, 1, lfields).GetRow(0)
				innerChk := genTestChunk(maxChunkSize, maxChunkSize, rfields)
				var defaultInner []types.Datum
				for i, f := range rfields {
					defaultInner = append(defaultInner, innerChk.GetRow(0).GetDatum(i, f))
				}
				joiner := NewJoiner(defaultCtx(), joinType, false, defaultInner, nil, lfields, rfields, nil, false)

				fields := make([]*types.FieldType, 0, len(lfields)+len(rfields))
				fields = append(fields, rfields...)
				fields = append(fields, lfields...)
				result := chunk.New(fields, maxChunkSize, maxChunkSize)

				for range 10 {
					required := rand.Int()%maxChunkSize + 1
					result.SetRequiredRows(required, maxChunkSize)
					result.Reset()
					it := chunk.NewIterator4Chunk(innerChk)
					it.Begin()
					_, _, err := joiner.TryToMatchInners(outerRow, it, result)
					require.NoError(t, err)
					require.Equal(t, required, result.NumRows())
				}
			}
		}
	}
}

func genTestChunk(maxChunkSize int, numRows int, fields []*types.FieldType) *chunk.Chunk {
	chk := chunk.New(fields, maxChunkSize, maxChunkSize)
	for numRows > 0 {
		numRows--
		for col, field := range fields {
			switch field.GetType() {
			case mysql.TypeLong:
				chk.AppendInt64(col, 0)
			case mysql.TypeFloat:
				chk.AppendFloat32(col, 0)
			default:
				panic("not support")
			}
		}
	}
	return chk
}

func TestIndexLookupJoinTaskPoolReset(t *testing.T) {
	outerTypes := []*types.FieldType{types.NewFieldType(mysql.TypeLong)}
	encodedTypes := []*types.FieldType{types.NewFieldType(mysql.TypeBlob)}
	e := &IndexLookUpJoin{}
	e.taskPool = sync.Pool{
		New: func() any {
			return &lookUpJoinTask{
				doneCh: make(chan error, 1),
				outerResult: chunk.NewList(
					outerTypes,
					1,
					4,
				),
				lookupMap:         mvmap.NewMVMap(),
				encodedLookUpKeys: []*chunk.Chunk{chunk.NewChunkWithCapacity(encodedTypes, 2)},
			}
		},
	}

	task := e.newLookUpJoinTask(nil)
	task.cursor = chunk.RowPtr{ChkIdx: 1, RowIdx: 2}
	task.hasMatch = true
	task.hasNull = true
	task.outerMatch = [][]bool{{true, false}}
	task.matchedInners = append(task.matchedInners, chunk.MutRowFromDatums([]types.Datum{types.NewIntDatum(1)}).ToRow())
	task.encodedLookUpKeys = task.encodedLookUpKeys[:1]
	task.encodedLookUpKeys[0].AppendBytes(0, []byte("stale"))
	task.lookupMap.Put([]byte("k"), []byte("v"))
	task.doneCh <- nil
	outerChk := chunk.New(outerTypes, 2, 2)
	outerChk.AppendInt64(0, 1)
	task.outerResult.Add(outerChk)

	e.recycleLookUpJoinTask(task)
	reused := e.newLookUpJoinTask(nil)
	require.Same(t, task, reused)
	require.Equal(t, chunk.RowPtr{}, reused.cursor)
	require.False(t, reused.hasMatch)
	require.False(t, reused.hasNull)
	require.Nil(t, reused.outerMatch)
	require.Empty(t, reused.matchedInners)
	require.Empty(t, reused.encodedLookUpKeys)
	require.Equal(t, 0, reused.outerResult.Len())
	require.Equal(t, 0, reused.lookupMap.Len())
	select {
	case <-reused.doneCh:
		t.Fatal("done channel should be drained before task reuse")
	default:
	}
}

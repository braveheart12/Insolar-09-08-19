//
// Copyright 2019 Insolar Technologies GmbH
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
//

package metrics_test

import (
	"math/rand"
	"net/http"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"

	"github.com/insolar/insolar/instrumentation/inslogger"
	"github.com/insolar/insolar/instrumentation/insmetrics"
	"github.com/insolar/insolar/ledger/storage/storagetest"
	"github.com/insolar/insolar/testutils/testmetrics"
)

func TestMetrics_NewMetrics(t *testing.T) {
	t.Parallel()
	ctx := inslogger.TestContext(t)
	testm := testmetrics.Start(ctx, t)

	var (
		// https://godoc.org/go.opencensus.io/stats
		videoCount = stats.Int64("example.com/measures/video_count", "number of processed videos", stats.UnitDimensionless)
		videoSize  = stats.Int64("video_size", "size of processed video", stats.UnitBytes)
	)
	osxtag := insmetrics.MustTagKey("osx")

	err := view.Register(
		&view.View{
			Measure:     videoCount,
			Aggregation: view.Count(),
			TagKeys:     []tag.Key{osxtag},
		},
		&view.View{
			Name:        "video_size",
			Measure:     videoSize,
			Aggregation: view.Distribution(0, 1<<16, 1<<32),
			TagKeys:     []tag.Key{osxtag},
		},
	)
	require.NoError(t, err)

	newctx := insmetrics.ChangeTags(ctx, tag.Insert(osxtag, "11.12.13"))
	stats.Record(newctx, videoCount.M(1), videoSize.M(rand.Int63()))

	content, err := testm.FetchContent()
	require.NoError(t, err)
	// fmt.Println("/metrics => ", content)

	assert.Regexp(t, regexp.MustCompile(`insolar_video_size_count{[^}]*osx="11\.12\.13"[^}]*} 1`), content )
	assert.Regexp(t, regexp.MustCompile(`insolar_example_com_measures_video_count{[^}]*osx="11\.12\.13"[^}]*} 1`), content)

	assert.NoError(t, testm.Stop())
}

func TestMetrics_ZPages(t *testing.T) {
	t.Parallel()
	ctx := inslogger.TestContext(t)
	testm := testmetrics.Start(ctx, t)

	// One more thing... from https://github.com/rakyll/opencensus-grpc-demo
	// also check /debug/rpcz
	code, content, err := testm.FetchURL("/debug/tracez")
	_ = content
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, code)
	// fmt.Println("/debug/tracez => ", content)

	assert.NoError(t, testm.Stop())
}

func TestMetrics_Badger(t *testing.T) {
	t.Parallel()
	ctx := inslogger.TestContext(t)

	_, cleaner := storagetest.TmpDB(ctx, nil)
	defer cleaner()

	testm := testmetrics.Start(ctx, t)

	code, content, err := testm.FetchURL("/metrics")
	_ = content
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, code)
	// fmt.Println("/metrics => ", content)
	assert.Contains(t, content, "insolar_badger_blocked_puts_total")

	assert.NoError(t, testm.Stop())
}

func TestMetrics_Status(t *testing.T) {
	t.Parallel()
	ctx := inslogger.TestContext(t)
	testm := testmetrics.Start(ctx, t)

	code, _, err := testm.FetchURL("/_status")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, code)

	assert.NoError(t, testm.Stop())
}

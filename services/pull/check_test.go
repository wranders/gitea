// Copyright 2019 The Gitea Authors.
// All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package pull

import (
	"context"
	"strconv"
	"testing"
	"time"

	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/modules/queue"

	"github.com/stretchr/testify/assert"
	"github.com/unknwon/com"
)

func TestPullRequest_AddToTaskQueue(t *testing.T) {
	assert.NoError(t, models.PrepareTestDatabase())

	idChan := make(chan int64, 10)

	q, err := queue.NewChannelUniqueQueue(func(data ...queue.Data) {
		for _, datum := range data {
			prID := datum.(string)
			id := com.StrTo(prID).MustInt64()
			idChan <- id
		}
	}, queue.ChannelUniqueQueueConfiguration{
		WorkerPoolConfiguration: queue.WorkerPoolConfiguration{
			QueueLength: 10,
			BatchLength: 1,
		},
		Workers: 1,
		Name:    "temporary-queue",
	}, "")
	assert.NoError(t, err)

	queueShutdown := []func(){}
	queueTerminate := []func(){}

	prQueue = q.(queue.UniqueQueue)

	pr := models.AssertExistsAndLoadBean(t, &models.PullRequest{ID: 2}).(*models.PullRequest)
	AddToTaskQueue(pr)

	assert.Eventually(t, func() bool {
		pr = models.AssertExistsAndLoadBean(t, &models.PullRequest{ID: 2}).(*models.PullRequest)
		return pr.Status == models.PullRequestStatusChecking
	}, 1*time.Second, 100*time.Millisecond)

	has, err := prQueue.Has(strconv.FormatInt(pr.ID, 10))
	assert.True(t, has)
	assert.NoError(t, err)

	prQueue.Run(func(_ context.Context, shutdown func()) {
		queueShutdown = append(queueShutdown, shutdown)
	}, func(_ context.Context, terminate func()) {
		queueTerminate = append(queueTerminate, terminate)
	})

	select {
	case id := <-idChan:
		assert.EqualValues(t, pr.ID, id)
	case <-time.After(time.Second):
		assert.Fail(t, "Timeout: nothing was added to pullRequestQueue")
	}

	has, err = prQueue.Has(strconv.FormatInt(pr.ID, 10))
	assert.False(t, has)
	assert.NoError(t, err)

	pr = models.AssertExistsAndLoadBean(t, &models.PullRequest{ID: 2}).(*models.PullRequest)
	assert.Equal(t, models.PullRequestStatusChecking, pr.Status)

	for _, callback := range queueShutdown {
		callback()
	}
	for _, callback := range queueTerminate {
		callback()
	}

	prQueue = nil
}

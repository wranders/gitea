// Copyright 2017 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package repository

import (
	"testing"

	"code.gitea.io/gitea/models"
	"github.com/stretchr/testify/assert"
)

func TestForkRepository(t *testing.T) {
	assert.NoError(t, models.PrepareTestDatabase())

	// user 13 has already forked repo10
	user := models.AssertExistsAndLoadBean(t, &models.User{ID: 13}).(*models.User)
	repo := models.AssertExistsAndLoadBean(t, &models.Repository{ID: 10}).(*models.Repository)

	fork, err := ForkRepository(user, user, repo, "test", "test")
	assert.Nil(t, fork)
	assert.Error(t, err)
	assert.True(t, models.IsErrForkAlreadyExist(err))
}

// Copyright 2017 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package models

import (
	"fmt"

	"xorm.io/builder"
)

// RepoIndexerType specifies the repository indexer type
type RepoIndexerType int

const (
	// RepoIndexerTypeCode code indexer
	RepoIndexerTypeCode RepoIndexerType = iota // 0
	// RepoIndexerTypeStats repository stats indexer
	RepoIndexerTypeStats // 1
)

// RepoIndexerStatus status of a repo's entry in the repo indexer
// For now, implicitly refers to default branch
type RepoIndexerStatus struct {
	ID          int64           `xorm:"pk autoincr"`
	RepoID      int64           `xorm:"INDEX(s)"`
	CommitSha   string          `xorm:"VARCHAR(40)"`
	IndexerType RepoIndexerType `xorm:"INDEX(s) NOT NULL DEFAULT 0"`
}

// GetUnindexedRepos returns repos which do not have an indexer status
func GetUnindexedRepos(indexerType RepoIndexerType, maxRepoID int64, page, pageSize int) ([]int64, error) {
	ids := make([]int64, 0, 50)
	cond := builder.Cond(builder.IsNull{
		"repo_indexer_status.id",
	})
	sess := x.Table("repository").Join("LEFT OUTER", "repo_indexer_status", "repository.id = repo_indexer_status.repo_id AND repo_indexer_status.indexer_type = ?", indexerType)
	if maxRepoID > 0 {
		cond = builder.And(cond, builder.Lte{
			"repository.id": maxRepoID,
		})
	}
	if page >= 0 && pageSize > 0 {
		start := 0
		if page > 0 {
			start = (page - 1) * pageSize
		}
		sess.Limit(pageSize, start)
	}

	sess.Where(cond).Cols("repository.id").Desc("repository.id")
	err := sess.Find(&ids)
	return ids, err
}

// getIndexerStatus loads repo codes indxer status
func (repo *Repository) getIndexerStatus(e Engine, indexerType RepoIndexerType) (*RepoIndexerStatus, error) {
	switch indexerType {
	case RepoIndexerTypeCode:
		if repo.CodeIndexerStatus != nil {
			return repo.CodeIndexerStatus, nil
		}
	case RepoIndexerTypeStats:
		if repo.StatsIndexerStatus != nil {
			return repo.StatsIndexerStatus, nil
		}
	}
	status := &RepoIndexerStatus{RepoID: repo.ID, IndexerType: indexerType}
	has, err := e.Get(status)
	if err != nil {
		return nil, err
	} else if !has {
		status.CommitSha = ""
	}
	switch indexerType {
	case RepoIndexerTypeCode:
		repo.CodeIndexerStatus = status
	case RepoIndexerTypeStats:
		repo.StatsIndexerStatus = status
	}
	return status, nil
}

// GetIndexerStatus loads repo codes indxer status
func (repo *Repository) GetIndexerStatus(indexerType RepoIndexerType) (*RepoIndexerStatus, error) {
	return repo.getIndexerStatus(x, indexerType)
}

// updateIndexerStatus updates indexer status
func (repo *Repository) updateIndexerStatus(e Engine, indexerType RepoIndexerType, sha string) error {
	status, err := repo.getIndexerStatus(e, indexerType)
	if err != nil {
		return fmt.Errorf("UpdateIndexerStatus: Unable to getIndexerStatus for repo: %s Error: %v", repo.FullName(), err)
	}

	if len(status.CommitSha) == 0 {
		status.CommitSha = sha
		_, err := e.Insert(status)
		if err != nil {
			return fmt.Errorf("UpdateIndexerStatus: Unable to insert repoIndexerStatus for repo: %s Sha: %s Error: %v", repo.FullName(), sha, err)
		}
		return nil
	}
	status.CommitSha = sha
	_, err = e.ID(status.ID).Cols("commit_sha").
		Update(status)
	if err != nil {
		return fmt.Errorf("UpdateIndexerStatus: Unable to update repoIndexerStatus for repo: %s Sha: %s Error: %v", repo.FullName(), sha, err)
	}
	return nil
}

// UpdateIndexerStatus updates indexer status
func (repo *Repository) UpdateIndexerStatus(indexerType RepoIndexerType, sha string) error {
	return repo.updateIndexerStatus(x, indexerType, sha)
}

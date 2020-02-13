// Copyright 2019 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import "gopkg.in/src-d/go-git.v4/plumbing/object"

// LastCommitCache cache
type LastCommitCache interface {
	Get(ref, entryPath string) (*object.Commit, error)
	Put(ref, entryPath, commitID string) error
}

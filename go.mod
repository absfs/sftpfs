module github.com/absfs/sftpfs

go 1.23

require (
	github.com/absfs/absfs v0.0.0-20251208232938-aa0ca30de832
	github.com/absfs/memfs v0.0.0-20251208230030-9f9671a4d047
	github.com/pkg/sftp v1.13.6
	golang.org/x/crypto v0.17.0
)

require (
	github.com/absfs/inode v0.0.0-20251208170702-9db24ab95ae4 // indirect
	github.com/kr/fs v0.1.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
)

replace (
	github.com/absfs/absfs => ../absfs
	github.com/absfs/fstesting => ../fstesting
	github.com/absfs/fstools => ../fstools
	github.com/absfs/inode => ../inode
	github.com/absfs/memfs => ../memfs
)

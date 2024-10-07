//go:build windows

/*
	"CreateFileSystemUnion"
	"EnumerateFileSystemUnions"
	"GetFileSystemUnionInformation"
	"MigrateFileSystemUnion"
	"RemoveFileSystemUnion"
	"RestoreFileSystemUnion"
	"RevertToLayerFileSystemUnion"
	"RevertToLayerFileSystemUnionTransacted"
	"SuppressFilesInFileSystemUnion"
	"SuppressFilesInFileSystemUnionTransacted"
*/

package winunionfs

import "syscall"

var (
	unionFSlib = syscall.NewLazyDLL("UnionFSApi.dll")

	UnionFScreateFileSystemUnion                    = unionFSlib.NewProc("CreateFileSystemUnion")
	UnionFSenumerateFileSystemUnions                = unionFSlib.NewProc("EnumerateFileSystemUnions")
	UnionFSgetFileSystemUnionInformation            = unionFSlib.NewProc("GetFileSystemUnionInformation")
	UnionFSmigrateFileSystemUnion                   = unionFSlib.NewProc("MigrateFileSystemUnion")
	UnionFSremoveFileSystemUnion                    = unionFSlib.NewProc("RemoveFileSystemUnion")
	UnionFSrestoreFileSystemUnion                   = unionFSlib.NewProc("RestoreFileSystemUnion")
	UnionFSrevertToLayerFileSystemUnion             = unionFSlib.NewProc("RevertToLayerFileSystemUnion")
	UnionFSrevertToLayerFileSystemUnionTransacted   = unionFSlib.NewProc("RevertToLayerFileSystemUnionTransacted")
	UnionFSsuppressFilesInFileSystemUnion           = unionFSlib.NewProc("SuppressFilesInFileSystemUnion")
	UnionFSsuppressFilesInFileSystemUnionTransacted = unionFSlib.NewProc("SuppressFilesInFileSystemUnionTransacted")
)

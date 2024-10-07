//go:build windows

// Create Virtual filesystem with Windows Projfs and merged folders
//
// To enable run with Admin in powershell: Enable-WindowsOptionalFeature -Online -FeatureName Client-ProjFS -NoRestart
//
// https://learn.microsoft.com/pt-br/windows/win32/projfs/projected-file-system
package overleyfs

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"unsafe"

	"github.com/google/uuid"
	"sirherobrine23.com.br/go-bds/go-bds/overleyfs/mergefs"
	"sirherobrine23.com.br/go-bds/go-bds/overleyfs/projfs"
)

const guidFile string = "_obgmgrproj.guid"

type enumerationSession struct {
	searchstr uintptr
	countget  int
	sentcount int
	wildcard  bool
}

type virtualizationStruct struct {
	enumerationsLocker *sync.Mutex
	enumerations       map[syscall.GUID]*enumerationSession
	_instanceHandle    projfs.PRJ_NAMESPACE_VIRTUALIZATION_CONTEXT
}

func CodeStatus(err error) uintptr {
	if err != nil {
		if os.IsNotExist(err) {
			return uintptr(0x80070002)
		} else if os.IsExist(err) {
			return uintptr(syscall.EEXIST)
		} else if os.IsPermission(err) {
			return uintptr(syscall.ENOENT)
		}
		return 1
	}
	return 0
}

func getVersionInfo(basicInfo *projfs.PRJ_FILE_BASIC_INFO) projfs.PRJ_PLACEHOLDER_VERSION_INFO {
	result := projfs.PRJ_PLACEHOLDER_VERSION_INFO{
		ProviderID: [projfs.PRJ_PLACEHOLDER_ID_LENGTH]byte{0, 0x1},
		ContentID:  [projfs.PRJ_PLACEHOLDER_ID_LENGTH]byte{0},
	}

	version := uint64(basicInfo.LastWriteTime.Nanoseconds())
	binary.LittleEndian.PutUint64(result.ContentID[:], version)
	return result
}

func FillInPlaceholderInfo(data *projfs.PRJ_PLACEHOLDER_INFO, fileinfo fs.FileInfo) {
	data.FileBasicInfo = toBasicInfo(fileinfo)
	data.VersionInfo = getVersionInfo(&data.FileBasicInfo)
}

// Unmount overlayfs
func (w *Overlayfs) Unmount() error {
	if virtStr, ok := w.internalStruct.(*virtualizationStruct); ok && virtStr._instanceHandle != 0 {
		projfs.PrjStopVirtualizing(virtStr._instanceHandle)
		virtStr._instanceHandle = 0
		return syscall.GetLastError()
	}
	return nil
}

// Mount overlayfs
func (w *Overlayfs) Mount() error {
	if instance, ok := w.internalStruct.(*virtualizationStruct); ok {
		if instance._instanceHandle != 0 {
			return fmt.Errorf("already started")
		}
	}

	if _, err := os.Stat(w.Target); os.IsNotExist(err) {
		if err = os.MkdirAll(w.Target, 0); err != nil {
			return err
		}
	}
	w.FS = mergefs.NewMergefsWithTopLayer(w.Upper, w.Lower...)

	var virtualizationGUID syscall.GUID
	if _, err := os.Stat(filepath.Join(w.Target, guidFile)); !os.IsNotExist(err) {
		b, err := os.ReadFile(filepath.Join(w.Target, guidFile))
		if err != nil {
			return err
		} else if len(b) != 16 {
			return fmt.Errorf("invalid GUID")
		}
		projfs.SetGUID(b, &virtualizationGUID)
	} else {
		uuid, err := uuid.NewRandom()
		if err != nil {
			return err
		}
		projfs.SetGUID(uuid[:], &virtualizationGUID)
		if err = os.WriteFile(filepath.Join(w.Target, guidFile), uuid[:], 0666); err != nil {
			return err
		}
	}

	if status := projfs.PrjMarkDirectoryAsPlaceholder(w.Target, "", nil, &virtualizationGUID); status != 0 {
		return fmt.Errorf("error to make directory placeholder, status code: 0x%08x", status)
	}

	options := &projfs.PRJ_STARTVIRTUALIZING_OPTIONS{
		PoolThreadCount:           0,
		ConcurrentThreadCount:     0,
		NotificationMappingsCount: 1,
		NotificationMappings: &projfs.PRJ_NOTIFICATION_MAPPING{
			NotificationRoot: projfs.GetPointer(""),
			NotificationBitMask: projfs.PRJ_NOTIFY_HARDLINK_CREATED |
				projfs.PRJ_NOTIFY_NEW_FILE_CREATED |
				projfs.PRJ_NOTIFY_FILE_OPENED |
				projfs.PRJ_NOTIFY_FILE_OVERWRITTEN |
				projfs.PRJ_NOTIFY_FILE_HANDLE_CLOSED_FILE_DELETED |
				projfs.PRJ_NOTIFY_FILE_HANDLE_CLOSED_FILE_MODIFIED,
		},
	}

	callback := &projfs.PRJ_CALLBACKS{
		NotificationCallback:              w.Notify,
		QueryFileNameCallback:             w.QueryFileName,
		CancelCommandCallback:             w.CancelCommand,
		StartDirectoryEnumerationCallback: w.StartDirectoryEnumeration,
		GetDirectoryEnumerationCallback:   w.GetDirectoryEnumeration,
		EndDirectoryEnumerationCallback:   w.EndDirectoryEnumeration,
		GetPlaceholderInfoCallback:        w.GetPlaceholderInfo,
		GetFileDataCallback:               w.GetFileData,
	}

	w.internalStruct = &virtualizationStruct{enumerationsLocker: &sync.Mutex{}, enumerations: make(map[syscall.GUID]*enumerationSession)}
	if status := projfs.PrjStartVirtualizing(w.Target, callback, w, options, &w.internalStruct.(*virtualizationStruct)._instanceHandle); status != 0 {
		return fmt.Errorf("cannot start folder virtualization, status code: 0x%08x", status)
	}
	return nil
}

func (instance *Overlayfs) Notify(callbackData *projfs.PRJ_CALLBACK_DATA, IsDirectory bool, notification projfs.PRJ_NOTIFICATION, destinationFileName uintptr, operationParameters *projfs.PRJ_NOTIFICATION_PARAMETERS) uintptr {
	// operation is done on file system
	filename := callbackData.GetFilePathName()
	fmt.Printf("Notify: %t %d %d '%s', %d\n", IsDirectory, callbackData.CommandId, notification, filename, *operationParameters)

	switch notification {
	case projfs.PRJ_NOTIFICATION_FILE_HANDLE_CLOSED_FILE_DELETED:
		return CodeStatus(instance.FS.Remove(filename))
	case projfs.PRJ_NOTIFICATION_HARDLINK_CREATED:
		return CodeStatus(instance.CreateHardLink(callbackData, destinationFileName))
	case projfs.PRJ_NOTIFICATION_NEW_FILE_CREATED:
		if IsDirectory {
			return CodeStatus(instance.FS.Mkdir(filename, 0777))
		}
		_, err := instance.FS.Create(filename)
		return CodeStatus(err)
	case projfs.PRJ_NOTIFICATION_FILE_HANDLE_CLOSED_FILE_MODIFIED, projfs.PRJ_NOTIFICATION_FILE_OVERWRITTEN:
		if !IsDirectory {
			file, err := os.Open(filepath.Join(instance.Target, filename))
			if err != nil {
				return CodeStatus(err)
			}
			defer file.Close()
			targetfile, err := instance.FS.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				return CodeStatus(err)
			}
			defer targetfile.Close()
			_, err = io.Copy(targetfile, file)
			return CodeStatus(err)
		}
	}

	return 0
}

func (instance *Overlayfs) CreateHardLink(callbackData *projfs.PRJ_CALLBACK_DATA, destinationFileName uintptr) error {
	filePath, targetPath := callbackData.GetFilePathName(), projfs.GetString(destinationFileName)
	if filePath == "" || targetPath == "" {
		return nil
	}
	return instance.FS.Link(filePath, targetPath)
}

func (instance *Overlayfs) QueryFileName(callbackData *projfs.PRJ_CALLBACK_DATA) uintptr { return 0 }
func (instance *Overlayfs) CancelCommand(callbackData *projfs.PRJ_CALLBACK_DATA) uintptr { return 0 }

func (instance *Overlayfs) StartDirectoryEnumeration(callbackData *projfs.PRJ_CALLBACK_DATA, enumerationId *syscall.GUID) uintptr {
	instance.internalStruct.(*virtualizationStruct).enumerationsLocker.Lock()
	defer instance.internalStruct.(*virtualizationStruct).enumerationsLocker.Unlock()
	instance.internalStruct.(*virtualizationStruct).enumerations[*enumerationId] = &enumerationSession{searchstr: 0, countget: 0, sentcount: 0, wildcard: false}
	return 0
}

func (instance *Overlayfs) EndDirectoryEnumeration(callbackData *projfs.PRJ_CALLBACK_DATA, enumerationId *syscall.GUID) uintptr {
	instance.internalStruct.(*virtualizationStruct).enumerationsLocker.Lock()
	defer instance.internalStruct.(*virtualizationStruct).enumerationsLocker.Unlock()
	delete(instance.internalStruct.(*virtualizationStruct).enumerations, *enumerationId)
	return 0
}

func (instance *Overlayfs) GetDirectoryEnumeration(callbackData *projfs.PRJ_CALLBACK_DATA, enumerationId *syscall.GUID, searchExpression uintptr, dirEntryBufferHandle projfs.PRJ_DIR_ENTRY_BUFFER_HANDLE) uintptr {
	instance.internalStruct.(*virtualizationStruct).enumerationsLocker.Lock()
	defer instance.internalStruct.(*virtualizationStruct).enumerationsLocker.Unlock()
	filenamepath := callbackData.GetFilePathName()
	first := instance.internalStruct.(*virtualizationStruct).enumerations[*enumerationId].countget == 0
	restart := callbackData.Flags&projfs.PRJ_CB_DATA_FLAG_ENUM_RESTART_SCAN != 0

	session, ok := instance.internalStruct.(*virtualizationStruct).enumerations[*enumerationId]
	if !ok {
		return uintptr(syscall.EINVAL)
	}

	if restart || first {
		session.sentcount = 0
		if searchExpression != 0 {
			session.searchstr = searchExpression
			session.wildcard = projfs.PrjDoesNameContainWildCards(searchExpression)
		} else {
			session.searchstr = 0
			session.wildcard = false
		}
	}
	instance.internalStruct.(*virtualizationStruct).enumerations[*enumerationId].countget++

	files, err := instance.FS.ReadDir(filenamepath)
	if err != nil {
		return uintptr(syscall.EIO)
	}

	for _, file := range files[session.sentcount:] {
		session.sentcount += 1
		fname := filepath.Base(file.Name())
		if strings.HasPrefix(fname, ".") || fname == guidFile {
			continue
		}

		if session.searchstr != 0 {
			match := projfs.PrjFileNameMatch(file.Name(), session.searchstr)
			if !match {
				continue
			}
		}
		info, err := file.Info()
		if err != nil {
			continue
		}
		dirEntry := toBasicInfo(info)
		projfs.PrjFillDirEntryBuffer(file.Name(), &dirEntry, dirEntryBufferHandle)
	}
	return 0
}

func toBasicInfo(file fs.FileInfo) projfs.PRJ_FILE_BASIC_INFO {
	ftime := syscall.NsecToFiletime(file.ModTime().UnixNano())
	return projfs.PRJ_FILE_BASIC_INFO{
		IsDirectory:    file.IsDir(),
		FileSize:       file.Size(),
		CreationTime:   ftime,
		LastAccessTime: ftime,
		LastWriteTime:  ftime,
		ChangeTime:     ftime,
		FileAttributes: 0,
	}
}

func (instance *Overlayfs) GetPlaceholderInfo(callbackData *projfs.PRJ_CALLBACK_DATA) uintptr {
	var data projfs.PRJ_PLACEHOLDER_INFO
	filename := callbackData.GetFilePathName()
	stat, err := instance.FS.Stat(filename)
	if os.IsNotExist(err) {
		return uintptr(0x80070002)
	} else if err != nil {
		return uintptr(syscall.EIO)
	}
	FillInPlaceholderInfo(&data, stat)
	return projfs.PrjWritePlaceholderInfo(instance.internalStruct.(*virtualizationStruct)._instanceHandle, callbackData.GetFilePathName(), &data, uint32(unsafe.Sizeof(data)))
}

func (instance *Overlayfs) GetFileData(callbackData *projfs.PRJ_CALLBACK_DATA, byteOffset uint64, length uint32) uintptr {
	filename := callbackData.GetFilePathName()
	file, err := instance.FS.Open(filename)
	if err != nil {
		return uintptr(syscall.EIO)
	}
	defer file.Close()
	buffer := make([]byte, length)

	var n int
	var count uint32
	for count < length {
		n, err = file.ReadAt(buffer[count:], int64(byteOffset+uint64(count)))
		count += uint32(n)
		if err == io.EOF {
			err = nil
			break
		}
	}

	if err != nil {
		return uintptr(syscall.EIO)
	}
	return projfs.PrjWriteFileData(instance.internalStruct.(*virtualizationStruct)._instanceHandle, &callbackData.DataStreamId, &buffer[0], byteOffset, length)
}

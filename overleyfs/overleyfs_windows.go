//go:build windows && cgo

// Create Virtual filesystem with Windows ProjFS and merged folders
package overleyfs

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"github.com/balazsgrill/potatodrive/win/projfs"
	"github.com/google/uuid"
	"sirherobrine23.com.br/go-bds/go-bds/overleyfs/mergefs"
)

type RemoteStateCache interface {
	UpdateHash(remotepath string, hash []byte) error
	GetHash(remotepath string) ([]byte, error)
}

type enumerationSession struct {
	searchstr uintptr
	countget  int
	sentcount int
	wildcard  bool
}

type virtualizationStruct struct {
	remoteCacheState RemoteStateCache
	_instanceHandle  projfs.PRJ_NAMESPACE_VIRTUALIZATION_CONTEXT
	enumerations     map[syscall.GUID]*enumerationSession
}

func SetGUID(b []byte, guid *syscall.GUID) {
	guid.Data1 = binary.LittleEndian.Uint32(b[0:4])
	guid.Data2 = binary.LittleEndian.Uint16(b[4:6])
	guid.Data3 = binary.LittleEndian.Uint16(b[6:8])
	guid.Data4 = ([8]byte)(b[8:16])
}

func GetPointer(str string) uintptr {
	ptr, err := syscall.UTF16PtrFromString(str)
	if err != nil {
		return 0
	}
	return uintptr(unsafe.Pointer(ptr))
}

// Stop Windows Project Filesystem virtualization
func (w *Overlayfs) Unmount() error {
	if virtStr, ok := w.internalStruct.(*virtualizationStruct); ok && virtStr._instanceHandle != 0{
		log.Printf("Stoping %v\n", virtStr._instanceHandle)
		projfs.PrjStopVirtualizing(virtStr._instanceHandle)
		virtStr._instanceHandle = 0
	}
	return nil
}

func (w *Overlayfs) Mount() error {
	if _, err := os.Stat(w.Target); os.IsNotExist(err) {
		if err = os.MkdirAll(w.Target, 0); err != nil {
			return err
		}
	}
	w.FS = mergefs.NewMergefsWithTopLayer(w.Upper, w.Lower...)

	var virtualizationGUID syscall.GUID
	if _, err := os.Stat(filepath.Join(w.Target, "_obgmgrproj.guid")); !os.IsNotExist(err) {
		b, err := os.ReadFile(filepath.Join(w.Target, "_obgmgrproj.guid"))
		if err != nil {
			return err
		} else if len(b) != 16 {
			return fmt.Errorf("invalid GUID")
		}
		SetGUID(b, &virtualizationGUID)
	} else {
		uuid, err := uuid.NewRandom()
		if err != nil {
			return err
		}
		SetGUID(uuid[:], &virtualizationGUID)
		if err = os.WriteFile(filepath.Join(w.Target, "_obgmgrproj.guid"), uuid[:], 0666); err != nil {
			return err
		}
	}

	status := projfs.PrjMarkDirectoryAsPlaceholder(w.Target, "", nil, &virtualizationGUID)
	if status != 0 {
		return fmt.Errorf("error to make directory placeholder, status code: 0x%08x", status)
	}

	options := &projfs.PRJ_STARTVIRTUALIZING_OPTIONS{
		NotificationMappingsCount: 1,
		PoolThreadCount:           4,
		ConcurrentThreadCount:     4,
		NotificationMappings: &projfs.PRJ_NOTIFICATION_MAPPING{
			NotificationBitMask: projfs.PRJ_NOTIFY_NEW_FILE_CREATED | projfs.PRJ_NOTIFY_FILE_OVERWRITTEN | projfs.PRJ_NOTIFY_FILE_HANDLE_CLOSED_FILE_DELETED | projfs.PRJ_NOTIFY_FILE_HANDLE_CLOSED_FILE_MODIFIED,
			NotificationRoot:    GetPointer(""),
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

	w.internalStruct = new(virtualizationStruct)
	status = projfs.PrjStartVirtualizing(w.Target, callback, w, options, &w.internalStruct.(*virtualizationStruct)._instanceHandle)
	if status != 0 {
		return fmt.Errorf("cannot start folder virtualization, status code: 0x%08x", status)
	}
	log.Printf("Starting virtualization Handler %v\n", w.internalStruct.(*virtualizationStruct)._instanceHandle)
	return nil
}

func returncode(err error) uintptr {
	if err != nil {
		return 1
	}
	return 0
}

func (instance *Overlayfs) Notify(callbackData *projfs.PRJ_CALLBACK_DATA, IsDirectory bool, notification projfs.PRJ_NOTIFICATION, destinationFileName uintptr, operationParameters *projfs.PRJ_NOTIFICATION_PARAMETERS) uintptr {
	// operation is done on file system
	filename := callbackData.GetFilePathName()
	log.Printf("Notify: %t %d %d '%s', %d", IsDirectory, callbackData.CommandId, notification, filename, *operationParameters)
	switch notification {

	case projfs.PRJ_NOTIFICATION_NEW_FILE_CREATED:
		if IsDirectory {
			return returncode(instance.FS.Mkdir(filename, 0777))
		} else {
			_, err := instance.FS.Create(filename)
			if err != nil {
				log.Print(err)
				return 1
			}
			return 0
		}
	case projfs.PRJ_NOTIFICATION_FILE_HANDLE_CLOSED_FILE_MODIFIED, projfs.PRJ_NOTIFICATION_FILE_OVERWRITTEN:
		if !IsDirectory {
			return returncode(instance.streamLocalToRemote(filename))
		}
	case projfs.PRJ_NOTIFICATION_FILE_HANDLE_CLOSED_FILE_DELETED:
		return returncode(instance.FS.Remove(filename))
	}
	return 0
}

func (instance *Overlayfs) streamLocalToRemote(filename string) error {
	file, err := instance.FS.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	data := make([]byte, 1024*1024)
	targetfile, err := instance.FS.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer targetfile.Close()

	hash := md5.New()
	for {
		n, err := file.Read(data)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		_, err = hash.Write(data[:n])
		if err != nil {
			return err
		}
		_, err = targetfile.Write(data[:n])
		if err != nil {
			return err
		}
	}

	return instance.internalStruct.(*virtualizationStruct).remoteCacheState.UpdateHash(filename, hash.Sum(nil))
}

func (instance *Overlayfs) QueryFileName(callbackData *projfs.PRJ_CALLBACK_DATA) uintptr {
	filename := callbackData.GetFilePathName()
	log.Printf("QueryFileName: '%s'", filename)
	return 0
}

func (instance *Overlayfs) CancelCommand(callbackData *projfs.PRJ_CALLBACK_DATA) uintptr {
	return 0
}

func (instance *Overlayfs) StartDirectoryEnumeration(callbackData *projfs.PRJ_CALLBACK_DATA, enumerationId *syscall.GUID) uintptr {
	log.Printf("StartDirectoryEnumeration: '%v'", *enumerationId)
	instance.internalStruct.(*virtualizationStruct).enumerations[*enumerationId] = &enumerationSession{
		searchstr: 0,
		countget:  0,
		sentcount: 0,
		wildcard:  false,
	}
	return 0
}

func (instance *Overlayfs) EndDirectoryEnumeration(callbackData *projfs.PRJ_CALLBACK_DATA, enumerationId *syscall.GUID) uintptr {
	log.Printf("EndDirectoryEnumeration: '%v'", *enumerationId)
	instance.internalStruct.(*virtualizationStruct).enumerations[*enumerationId] = nil
	return 0
}

func (instance *Overlayfs) GetDirectoryEnumeration(callbackData *projfs.PRJ_CALLBACK_DATA, enumerationId *syscall.GUID, searchExpression uintptr, dirEntryBufferHandle projfs.PRJ_DIR_ENTRY_BUFFER_HANDLE) uintptr {
	filenamepath := callbackData.GetFilePathName()
	first := instance.internalStruct.(*virtualizationStruct).enumerations[*enumerationId].countget == 0
	restart := callbackData.Flags&projfs.PRJ_CB_DATA_FLAG_ENUM_RESTART_SCAN != 0

	session, ok := instance.internalStruct.(*virtualizationStruct).enumerations[*enumerationId]
	if !ok {
		return uintptr(syscall.EINVAL)
	}
	log.Printf("GetDirectoryEnumeration (%t, %t, %d) %s", first, restart, session.sentcount, filenamepath)

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
		log.Printf("Error reading directory %s: %s", filenamepath, err)
		return uintptr(syscall.EIO)
	}

	for _, file := range files[session.sentcount:] {
		session.sentcount += 1
		fname := filepath.Base(file.Name())
		if strings.HasPrefix(fname, ".") {
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
	log.Printf("Sent %d entries", session.sentcount)
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

func (instance *Overlayfs) GetPlaceholderInfo(callbackData *projfs.PRJ_CALLBACK_DATA) uintptr {
	var data projfs.PRJ_PLACEHOLDER_INFO
	filename := callbackData.GetFilePathName()
	log.Printf("GetPlaceholderInfo %s", filename)
	stat, err := instance.FS.Stat(filename)
	if os.IsNotExist(err) {
		return uintptr(0x80070002)
	}
	if err != nil {
		log.Printf("Error getting placeholder info for %s: %s", filename, err)
		return uintptr(syscall.EIO)
	}
	FillInPlaceholderInfo(&data, stat)
	return projfs.PrjWritePlaceholderInfo(instance.internalStruct.(*virtualizationStruct)._instanceHandle, callbackData.GetFilePathName(), &data, uint32(unsafe.Sizeof(data)))
}

func (instance *Overlayfs) GetFileData(callbackData *projfs.PRJ_CALLBACK_DATA, byteOffset uint64, length uint32) uintptr {
	filename := callbackData.GetFilePathName()
	log.Printf("GetFileData %s[%d]@%d", filename, length, byteOffset)
	file, err := instance.FS.Open(filename)
	if err != nil {
		log.Printf("Error opening file %s: %s", filename, err)
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

	log.Printf("Read %d bytes", count)
	if err != nil {
		log.Printf("Error reading file %s: %s", filename, err)
		return uintptr(syscall.EIO)
	}
	return projfs.PrjWriteFileData(instance.internalStruct.(*virtualizationStruct)._instanceHandle, &callbackData.DataStreamId, &buffer[0], byteOffset, length)
}

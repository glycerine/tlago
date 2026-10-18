package tlc

import (
	"io"
	"os"
)

func ioUtilsTXTReadFile(path string) (data []byte, err error) {
	file, err := os.Open(ioUtilsTXTSystemPath(path))
	if err != nil {
		return nil, ioUtilsTXTPathError(path, err)
	}
	defer func() {
		failure := recover()
		closeErr := file.Close()
		if closeErr != nil {
			closeErr = ioUtilsTXTChannelError(closeErr)
		}
		if failure != nil {
			if primary, ok := failure.(error); ok {
				javaSuppressCloseError(primary, closeErr)
			}
			panic(failure)
		}
		if err == nil {
			if closeErr != nil {
				data, err = nil, closeErr
			}
		} else {
			javaSuppressCloseError(err, closeErr)
		}
	}()
	info, err := file.Stat()
	if err != nil {
		return nil, ioUtilsTXTChannelError(err)
	}
	if info.Size() > 0x7fffffff {
		return nil, NewOutOfMemoryError("Required array size too large")
	}
	data, err = io.ReadAll(file)
	if err != nil {
		return nil, ioUtilsTXTChannelError(err)
	}
	return data, nil
}

func ioUtilsTXTWriteFile(path string, data []byte, options ioUtilsFileOptions) (err error) {
	file, err := ioUtilsTXTOpen(path, options)
	if err != nil {
		return err
	}
	defer func() {
		failure := recover()
		closeErr := file.Close()
		if closeErr != nil {
			closeErr = ioUtilsTXTChannelError(closeErr)
		}
		if failure != nil {
			ioUtilsTXTDeleteAfterClose(path, options)
			if primary, ok := failure.(error); ok {
				javaSuppressCloseError(primary, closeErr)
			}
			panic(failure)
		}
		if err == nil {
			err = closeErr
		} else {
			javaSuppressCloseError(err, closeErr)
		}
		ioUtilsTXTDeleteAfterClose(path, options)
	}()
	// Files.write completes chunks of at most 8192 bytes in its channel stream.
	for len(data) > 0 {
		chunk := data
		if len(chunk) > 8192 {
			chunk = chunk[:8192]
		}
		if _, err := file.Write(chunk); err != nil {
			return ioUtilsTXTChannelError(err)
		}
		data = data[len(chunk):]
	}
	return nil
}

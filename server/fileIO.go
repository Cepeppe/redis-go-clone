package main

import (
	"encoding/binary"
	"os"
)


var NATIVE_ENDIAN = binary.NativeEndian

// An entry in: key_len(uint_32) key(string) data_len(uint_32)
// returns key and value
func read_entry(f *os.File) (string, string, error){

	// READ KEY LEN
	var keyLen uint32

	// err is io.ErrUnexpectedEOF if file ended before N bytes,
	// or io.EOF if the file had 0 bytes left.
	if err := binary.Read(f, NATIVE_ENDIAN, &keyLen); err != nil {
		return "", "", err
	}

	// READ KEY
	key_buf := make([]byte, keyLen)
	if err := binary.Read(f, NATIVE_ENDIAN, &key_buf); err != nil {
		return "", "", err
	}

	// READ DATA LEN
	var data_len uint32
	if err := binary.Read(f, NATIVE_ENDIAN, &data_len); err != nil {
		return "", "", err
	}

	// READ DATA
	data_buf := make([]byte, data_len)
	if err := binary.Read(f, NATIVE_ENDIAN, data_buf); err != nil {
		return "", "", err
	}

	return string(key_buf), string(data_buf), nil
}


// An entry in: key_len(uint_32) key(string) data_len(uint_32)
// returns key and value
func write_entry(f *os.File, key string, data string) error{
	
	var err = binary.Write(f, NATIVE_ENDIAN, uint32(len(key)))
	if err != nil { return err }
	
	_, err = f.Write([]byte(key))
	if err != nil { return err }

	err = binary.Write(f, NATIVE_ENDIAN, uint32(len(data)))
	if err != nil { return err }

	_, err = f.Write([]byte(data))
	if err != nil { return err }

	return nil
}



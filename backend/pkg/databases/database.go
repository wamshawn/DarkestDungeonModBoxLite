package databases

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"

	"github.com/rs/xid"
	"github.com/tidwall/buntdb"
)

const (
	versionKey = "$version"
)

type ListEntry struct {
	Index uint `json:"index"`
}

type Index struct {
	Name    string
	Pattern string
	Less    []func(a, b string) bool
}

func CreateIndex(name, pattern string, less ...func(a, b string) bool) Index {
	return Index{
		Name:    name,
		Pattern: pattern,
		Less:    less,
	}
}

func New(filename string, indexes ...Index) (db *Database, err error) {
	kv, kvErr := buntdb.Open(filename)
	if kvErr != nil {
		err = kvErr
		return
	}
	db = &Database{
		file:    filename,
		kv:      kv,
		indexes: indexes,
	}
	if err = db.createIndexes(); err != nil {
		_ = kv.Close()
		return
	}
	return
}

type Database struct {
	file    string
	kv      *buntdb.DB
	indexes []Index
}

func (db *Database) Load(filename string) (err error) {
	b, bErr := os.ReadFile(filename)
	if bErr != nil {
		err = bErr
		return
	}
	if len(b) == 0 {
		return
	}
	kv, kvErr := buntdb.Open(filename)
	if kvErr != nil {
		err = kvErr
		return
	}
	_ = kv.Close()

	_ = db.kv.Close()

	if err = os.WriteFile(db.file, b, 0644); err != nil {
		return
	}
	kv, kvErr = buntdb.Open(db.file)
	if kvErr != nil {
		err = kvErr
		return
	}
	db.kv = kv
	if err = db.createIndexes(); err != nil {
		return
	}
	return
}

func (db *Database) Save(filename string) (err error) {
	buf := bytes.NewBuffer(nil)
	if err = db.kv.Save(buf); err != nil {
		return
	}
	err = os.WriteFile(filename, buf.Bytes(), 0644)
	return
}

func (db *Database) createIndexes() (err error) {
	for _, index := range db.indexes {
		if err = db.kv.CreateIndex(index.Name, index.Pattern, index.Less...); err != nil {
			return
		}
	}
	return
}

func (db *Database) Version() uint64 {
	tx, txErr := db.kv.Begin(false)
	if txErr != nil {
		return 0
	}
	defer tx.Rollback()
	versionStr, versionStrErr := tx.Get(versionKey)
	if versionStrErr != nil {
		return 0
	}
	if versionStr != "" {
		version, err := strconv.ParseUint(versionStr, 16, 64)
		if err != nil {
			return 0
		}
		return version
	}
	return 0
}

func (db *Database) IncrVersion() {
	version := db.Version()
	version++
	vs := strconv.FormatUint(version, 16)
	_ = db.kv.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(versionKey, vs, nil)
		return err
	})
}

func (db *Database) Update(key string, value any) (err error) {
	b, bErr := json.Marshal(value)
	if bErr != nil {
		err = bErr
		return
	}
	return db.kv.Update(func(tx *buntdb.Tx) error {
		_, _, setErr := tx.Set(key, string(b), nil)
		return setErr
	})
}

func (db *Database) Remove(key string) (err error) {
	return db.kv.Update(func(tx *buntdb.Tx) error {
		_, deleteErr := tx.Delete(key)
		return deleteErr
	})
}

func (db *Database) Get(key string, value any) (has bool, err error) {
	err = db.kv.View(func(tx *buntdb.Tx) error {
		val, getErr := tx.Get(key)
		if getErr != nil {
			return getErr
		}
		return json.Unmarshal([]byte(val), value)
	})
	if err != nil {
		if errors.Is(err, buntdb.ErrNotFound) {
			err = nil
		}
		return
	}
	has = true
	return
}

func (db *Database) AscendKey(key string, values any, less ...func(string, string) bool) (err error) {
	tx, txErr := db.kv.Begin(true)
	if txErr != nil {
		err = txErr
		return
	}
	defer tx.Rollback()
	index := xid.New().String()
	if err = tx.CreateIndex(index, fmt.Sprintf("%s:*", key), less...); err != nil {
		return
	}
	defer tx.DropIndex(index)

	buf := bytes.NewBuffer(nil)
	buf.WriteByte('[')
	i := 0
	err = tx.Ascend(index, func(key, value string) bool {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(value)
		return true
	})
	if err != nil {
		return
	}
	buf.WriteByte(']')
	if buf.Len() == 2 {
		return
	}
	err = json.Unmarshal(buf.Bytes(), &values)
	return
}

func (db *Database) AscendKeys(pattern string, values any) (err error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte('[')
	err = db.kv.View(func(tx *buntdb.Tx) error {
		i := 0
		return tx.AscendKeys(pattern, func(key, value string) bool {
			if i > 0 {
				buf.WriteByte(',')
			}
			buf.WriteString(value)
			return true
		})
	})
	if err != nil {
		return
	}
	buf.WriteByte(']')
	if buf.Len() == 2 {
		return
	}
	err = json.Unmarshal(buf.Bytes(), &values)
	return
}

func (db *Database) Ascend(index string, values any) (err error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte('[')
	err = db.kv.View(func(tx *buntdb.Tx) error {
		i := 0
		return tx.Ascend(index, func(key, value string) bool {
			if i > 0 {
				buf.WriteByte(',')
			}
			buf.WriteString(value)
			return true
		})
	})
	if err != nil {
		return
	}
	buf.WriteByte(']')
	if buf.Len() == 2 {
		return
	}
	err = json.Unmarshal(buf.Bytes(), &values)
	return
}

func (db *Database) Range(index string, beg uint, end uint, values any) (err error) {
	rt := reflect.TypeOf(values)
	if rt.Kind() != reflect.Ptr {
		err = errors.New("values must be a pointer")
		return
	}
	rt = rt.Elem()
	if rt.Kind() != reflect.Slice {
		err = errors.New("values element must be a slice")
		return
	}
	rt = rt.Elem()
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	ft, has := rt.FieldByName("Index")
	if !has {
		err = errors.New("no index field")
		return
	}
	if ft.Type != reflect.TypeOf(uint(0)) {
		err = errors.New("index field must be an uint")
		return
	}
	tag := ft.Tag.Get("json")
	if tag == "" {
		tag = ft.Name
	}
	begFlag := fmt.Sprintf("{\"%s\": %d}", tag, beg)
	endFlag := fmt.Sprintf("{\"%s\": %d}", tag, end)
	buf := bytes.NewBuffer(nil)
	buf.WriteByte('[')
	err = db.kv.View(func(tx *buntdb.Tx) error {
		i := 0
		return tx.AscendRange(index, begFlag, endFlag, func(key, value string) bool {
			if i > 0 {
				buf.WriteByte(',')
			}
			buf.WriteString(value)
			return true
		})
	})
	if err != nil {
		return
	}
	buf.WriteByte(']')
	if buf.Len() == 2 {
		return
	}
	err = json.Unmarshal(buf.Bytes(), &values)
	return
}

func (db *Database) Close() {
	_ = db.kv.Close()
}

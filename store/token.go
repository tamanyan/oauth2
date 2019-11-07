package store

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"

	// _ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"

	"github.com/tamanyan/oauth2"
	"github.com/tamanyan/oauth2/models"
)

var noUpdateContent = "No content found to be updated"
var mutex = &sync.Mutex{}

// StoreItem data item saved into db
type StoreItem struct {
	gorm.Model
	//ID        int64 `gorm:"AUTO_INCREMENT"`
	ExpiredAt int64
	Code      string `gorm:"type:varchar(512)"`
	Access    string `gorm:"type:varchar(512)"`
	Refresh   string `gorm:"type:varchar(512)"`
	Data      string `gorm:"type:text"`
}

// NewTokenConfig create mysql configuration instance
func NewTokenConfig(dsn string, dbType string, tableName string) *TokenConfig {
	return &TokenConfig{
		DSN:         dsn,
		DBType:      dbType,
		TableName:   tableName,
		MaxLifetime: time.Hour * 2,
	}
}

// TokenConfig xorm configuration
type TokenConfig struct {
	DSN         string
	DBType      string
	TableName   string
	MaxLifetime time.Duration
}

// NewTokenStore create mysql store instance,
func NewTokenStore(config *TokenConfig, gcInterval int) (store oauth2.TokenStore, err error) {
	db, err := gorm.Open(config.DBType, config.DSN)
	if err != nil {
		return
	}
	return NewStoreWithDB(config, db, gcInterval)
}

// NewStoreWithDB create store with config
func NewStoreWithDB(config *TokenConfig, db *gorm.DB, gcInterval int) (store oauth2.TokenStore, err error) {
	tokenStore := &Store{
		db:        db,
		tableName: "oauth2_token",
		stdout:    os.Stderr,
	}
	if config.TableName != "" {
		tokenStore.tableName = config.TableName
	}
	interval := 600
	if gcInterval > 0 {
		interval = gcInterval
	}
	tokenStore.ticker = time.NewTicker(time.Second * time.Duration(interval))

	if !db.HasTable(tokenStore.tableName) {
		if err := db.Table(tokenStore.tableName).CreateTable(&StoreItem{}).Error; err != nil {
			panic(err)
		}
	}

	go tokenStore.gc()
	store = tokenStore

	return
}

// Store mysql token store
type Store struct {
	tableName string
	db        *gorm.DB
	stdout    io.Writer
	ticker    *time.Ticker
}

// SetStdout set error output
func (s *Store) SetStdout(stdout io.Writer) *Store {
	s.stdout = stdout
	return s
}

// Close close the store
func (s *Store) Close() {
	s.ticker.Stop()
}

func (s *Store) errorf(format string, args ...interface{}) {
	if s.stdout != nil {
		buf := fmt.Sprintf(format, args...)
		s.stdout.Write([]byte(buf))
	}
}

func (s *Store) gc() {
	for range s.ticker.C {
		now := time.Now().Unix()
		var count int
		if err := s.db.Table(s.tableName).Where("expired_at > ?", now).Count(&count).Error; err != nil {
			s.errorf("[ERROR]:%s\n", err)
			return
		}
		s.errorf("DELETE:%d\n", count)
		if count > 0 {
			if err := s.db.Table(s.tableName).Where("expired_at > ?", now).Delete(&StoreItem{}).Error; err != nil {
				s.errorf("[ERROR]:%s\n", err)
			} else {
				s.errorf("DELETE:%d\n", count)
			}
		}
	}
}

// Create create and store the new token information
func (s *Store) Create(info oauth2.TokenInfo) error {
	jv, err := json.Marshal(info)
	if err != nil {
		return err
	}
	item := &StoreItem{
		Data: string(jv),
	}

	if code := info.GetCode(); code != "" {
		item.Code = code
		item.ExpiredAt = info.GetCodeCreateAt().Add(info.GetCodeExpiresIn()).Unix()
	} else {
		item.Access = info.GetAccess()
		item.ExpiredAt = info.GetAccessCreateAt().Add(info.GetAccessExpiresIn()).Unix()

		if refresh := info.GetRefresh(); refresh != "" {
			item.Refresh = info.GetRefresh()
			item.ExpiredAt = info.GetRefreshCreateAt().Add(info.GetRefreshExpiresIn()).Unix()
		}
	}

	err = s.db.Table(s.tableName).Create(item).Error
	return err
}

// RemoveByCode delete the authorization code
func (s *Store) RemoveByCode(code string) error {
	return s.db.Table(s.tableName).Where("code = ?", code).Update("code", "").Error
}

// RemoveByAccess use the access token to delete the token information
func (s *Store) RemoveByAccess(access string) error {
	return s.db.Table(s.tableName).Where("access = ?", access).Update("access", "").Error
}

// RemoveByRefresh use the refresh token to delete the token information
func (s *Store) RemoveByRefresh(refresh string) error {
	return s.db.Table(s.tableName).Where("refresh = ?", refresh).Update("refresh", "").Error
}

func (s *Store) toTokenInfo(data string) (oauth2.TokenInfo, error) {
	var tm models.Token
	err := json.Unmarshal([]byte(data), &tm)
	return &tm, err
}

// GetByCode use the authorization code for token information data
func (s *Store) GetByCode(code string) (oauth2.TokenInfo, error) {
	if code == "" {
		return nil, nil
	}

	var item StoreItem
	if err := s.db.Table(s.tableName).Where("code = ?", code).Find(&item).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return s.toTokenInfo(item.Data)
}

// GetByAccess use the access token for token information data
func (s *Store) GetByAccess(access string) (oauth2.TokenInfo, error) {
	if access == "" {
		return nil, nil
	}

	var item StoreItem
	if err := s.db.Table(s.tableName).Where("access = ?", access).Find(&item).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return s.toTokenInfo(item.Data)
}

// GetByRefresh use the refresh token for token information data
func (s *Store) GetByRefresh(refresh string) (oauth2.TokenInfo, error) {
	if refresh == "" {
		return nil, nil
	}

	var item StoreItem
	if err := s.db.Table(s.tableName).Where("refresh = ?", refresh).Find(&item).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return s.toTokenInfo(item.Data)
}

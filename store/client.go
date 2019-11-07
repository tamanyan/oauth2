package store

import (
	// "errors"
	// "sync"
	"io"
	"time"
	// "fmt"
	// "log"

	"github.com/jinzhu/gorm"

	"github.com/tamanyan/oauth2"
	"github.com/tamanyan/oauth2/models"
)

// ClientItem data item
type ClientItem struct {
	//ID        int64 `gorm:"AUTO_INCREMENT"`
	ID        string `gorm:"type:varchar(512)"`
	Secret    string `gorm:"type:varchar(512)"`
	Domain    string `gorm:"type:varchar(512)"`
	UserID    string `gorm:"type:varchar(512)"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
}

// NewClientConfig create mysql configuration instance
func NewClientConfig(dsn string, dbType string, tableName string) *ClientConfig {
	return &ClientConfig{
		DSN:       dsn,
		DBType:    dbType,
		TableName: tableName,
	}
}

// ClientConfig xorm configuration
type ClientConfig struct {
	DSN       string
	DBType    string
	TableName string
}

// NewClientStore create a token store instance based on memory
func NewClientStore(config *ClientConfig) (client *ClientStore, err error) {
	db, err := gorm.Open(config.DBType, config.DSN)
	if err != nil {
		return
	}
	return NewClientWithDB(config, db)
}

// NewClientWithDB create client store
func NewClientWithDB(config *ClientConfig, db *gorm.DB) (client *ClientStore, err error) {
	client = &ClientStore{
		tableName: "client",
		db:        db,
	}

	if config.TableName != "" {
		client.tableName = config.TableName
	}

	if !db.HasTable(client.tableName) {
		if err := db.Table(client.tableName).CreateTable(&ClientItem{}).Error; err != nil {
			panic(err)
		}
	}

	return
}

// ClientStore client information store
type ClientStore struct {
	tableName string
	db        *gorm.DB
	stdout    io.Writer
}

// GetByID according to the ID for the client information
func (cs *ClientStore) GetByID(id string) (cli oauth2.ClientInfo, err error) {
	if id == "" {
		return nil, nil
	}

	var item ClientItem
	if err := cs.db.Table(cs.tableName).Where("id = ?", id).Find(&item).Error; err != nil {
		return nil, err
	}

	return cs.toClientInfo(&item), nil
}

// Set set client information
func (cs *ClientStore) Set(id string, cli oauth2.ClientInfo) (err error) {
	if id != cli.GetID() || id == "" {
		return nil
	}

	item := &ClientItem{
		ID:     cli.GetID(),
		Secret: cli.GetSecret(),
		Domain: cli.GetDomain(),
		UserID: cli.GetUserID(),
	}

	return cs.db.Table(cs.tableName).Save(item).Error
}

func (cs *ClientStore) toClientInfo(item *ClientItem) oauth2.ClientInfo {
	var client = &models.Client{
		ID:     item.ID,
		Secret: item.Secret,
		Domain: item.Domain,
		UserID: item.UserID,
	}
	return client
}

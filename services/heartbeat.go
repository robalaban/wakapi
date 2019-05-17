package services

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/n1try/wakapi/models"
	gormbulk "github.com/t-tiger/gorm-bulk-insert"
)

const TableHeartbeat = "heartbeat"

type HeartbeatService struct {
	Cofnig *models.Config
	Db     *gorm.DB
}

func (srv *HeartbeatService) InsertBatch(heartbeats []*models.Heartbeat) error {
	var batch []interface{}
	for _, h := range heartbeats {
		batch = append(batch, h)
	}

	if err := gormbulk.BulkInsert(srv.Db, batch, 3000); err != nil {
		return err
	}
	return nil
}

func (srv *HeartbeatService) GetAllFrom(date time.Time, user *models.User) ([]*models.Heartbeat, error) {
	var heartbeats []*models.Heartbeat
	if err := srv.Db.
		Where(&models.Heartbeat{UserID: user.ID}).
		Where("time > ?", date).
		Find(heartbeats).Error; err != nil {
		return nil, err
	}
	return heartbeats, nil
}

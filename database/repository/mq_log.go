package repository

import (
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/mqlog"
)

func (r *Repository) GetQueuePosition(messageId uuid.UUID) (int, error) {
	// Figure out where this message lands based on priority and created_at
	query := `
	SELECT row_num
	FROM (
			SELECT message_id,
						 ROW_NUMBER() OVER (ORDER BY priority DESC, created_at ASC) as row_num
			FROM mq_log
	) AS subquery
	WHERE message_id = $1
	`

	rows, err := r.DB.QueryContext(r.Ctx, query, messageId)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var position int
	if rows.Next() {
		err = rows.Scan(&position)
		if err != nil {
			return 0, err
		}
	} else {
		// No rows means not in the queue
		return 0, nil
	}

	// Check for any errors from iterating over rows.
	if err = rows.Err(); err != nil {
		return 0, err
	}

	return position, nil
}

// Add to queue log
func (r *Repository) AddToQueueLog(messageId uuid.UUID, priority int, DB *ent.Client) (*ent.MqLog, error) {
	if DB == nil {
		DB = r.DB
	}
	return DB.MqLog.Create().SetMessageID(messageId).SetPriority(priority).Save(r.Ctx)
}

// Delete from queue log
func (r *Repository) DeleteFromQueueLog(messageId uuid.UUID, DB *ent.Client) (int, error) {
	if DB == nil {
		DB = r.DB
	}
	return DB.MqLog.Delete().Where(mqlog.MessageIDEQ(messageId)).Exec(r.Ctx)
}

// Set is_processing
func (r *Repository) SetIsProcessingInQueueLog(messageId uuid.UUID, isProcessing bool, DB *ent.Client) (int, error) {
	if DB == nil {
		DB = r.DB
	}
	return DB.MqLog.Update().Where(mqlog.MessageIDEQ(messageId)).SetIsProcessing(isProcessing).Save(r.Ctx)
}

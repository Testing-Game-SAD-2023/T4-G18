package robot

import (
	"fmt"
	"math/rand"

	"github.com/alarmfox/game-repository/model"
	"github.com/alarmfox/game-repository/web"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RobotStorage struct {
	db *gorm.DB
}

func NewRobotStorage(db *gorm.DB) *RobotStorage {
	return &RobotStorage{
		db: db,
	}
}

func (rs *RobotStorage) CreateBulk(r *CreateRequest) (int, error) {
	robots := make([]model.Robot, len(r.Robots))

	for i, robot := range r.Robots {
		robots[i] = model.Robot{
			TestClassId: robot.TestClassId,
			Scores:      robot.Scores,
			Difficulty:  robot.Difficulty,
			Type:        int8(robot.Type),
		}
	}

	err := rs.db.
		Clauses(clause.OnConflict{
			UpdateAll: true,
		}).
		CreateInBatches(&robots, 100).
		Error

	return len(robots), web.MakeServiceError(err)
}

func (gs *RobotStorage) FindByFilter(testClassId string, difficulty string, t RobotType) (Robot, error) {
	var (
		robot model.Robot
		ids   []int64
	)

	err := gs.db.Transaction(func(tx *gorm.DB) error {
		err := tx.
			Model(&model.Robot{}).
			Select("id").
			Where(&model.Robot{
				TestClassId: testClassId,
				Difficulty:  difficulty,
			}).
			Where("type = ? ", t).
			Find(&ids).
			Error

		if err != nil {
			return err
		}
		if len(ids) == 0 {
			return gorm.ErrRecordNotFound
		}
		var id int64
		switch t {
		case evosuite:
			id = ids[0]
		case randoop:
			pos := rand.Intn(len(ids))
			id = ids[pos]
		default:
			return fmt.Errorf("%w: unsupported test engine", web.ErrInvalidParam)
		}

		return tx.First(&robot, id).Error

	})

	return *fromModel(&robot), web.MakeServiceError(err)
}

func (rs *RobotStorage) DeleteByTestClass(testClassId string) error {

	db := rs.db.Where(&model.Robot{TestClassId: testClassId}).
		Delete(&[]model.Robot{})
	if db.Error != nil {
		return db.Error
	} else if db.RowsAffected < 1 {
		return web.ErrNotFound
	}

	return nil
}

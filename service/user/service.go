package user

import (
	"github.com/gopl-dev/server/ds"
)

func Create(u *ds.User) (err error) {
	//err = database.ORM().Insert(u)
	return
}

func Update(u *ds.User) (err error) {
	//err = database.ORM().Update(u)
	return
}

func SetActiveAt(userID int64) (err error) {
	//_, err = database.ORM().
	//	Model(&model.User{}).
	//	Where("id = ?", userID).
	//	Set("active_at = ?", time.Now()).
	//	Update()

	return
}

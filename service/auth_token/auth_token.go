package authtoken

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/gopl-dev/server/ds"
)

func Create(elem *ds.AuthToken) (err error) {
	//elem.Token = fmt.Sprintf("%x", sha256.Sum256([]byte(util.RandomString(32))))

	//_, err = database.ORM().
	//	Model(elem).
	//	Insert()

	return
}

func FindByToken(token string) (t ds.AuthToken, err error) {
	//err = database.ORM().
	//	Model(&t).
	//	Relation("User").
	//	Where("token = ?", token).
	//	First()
	//if err == pg.ErrNoRows {
	//	err = errors.New("auth token not found")
	//}

	return
}

func Sign(m *ds.AuthToken) string {
	data := []string{
		fmt.Sprintf("%d", m.ID),
		fmt.Sprintf("%d", m.UserID),
		m.ClientName,
		m.UserAgent,
		m.Token,
		m.CreatedAt.UTC().Format(time.RFC3339),
	}

	sig := fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(data, "-"))))
	return m.Token + sig
}

func DeleteByUser(id string) (err error) {
	//_, err = database.ORM().
	//	Model(&model.AuthToken{}).
	//	Where("user_id = ?", id).
	//	Delete()

	return
}

func Prolong(id int64) (err error) {
	//_, err = database.ORM().
	//	Model(&model.AuthToken{}).
	//	Where("id = ?", id).
	//	Set("expires_at = ?", time.Now().Add(config.Get().AuthTokenLifetime)).
	//	Update()

	return
}

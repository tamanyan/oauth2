package store_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/tamanyan/oauth2/models"
	"github.com/tamanyan/oauth2/store"
)

func TestClientStore(t *testing.T) {
	Convey("Test client store", t, func() {
		clientStore, err := store.NewClientStore(store.NewClientConfig(":memory:", "sqlite3", "client"))

		err = clientStore.Set("1", &models.Client{ID: "1", Secret: "2"})
		So(err, ShouldBeNil)

		cli, err := clientStore.GetByID("1")
		So(err, ShouldBeNil)
		So(cli.GetID(), ShouldEqual, "1")
	})
}

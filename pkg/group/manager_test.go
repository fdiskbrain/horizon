package group

import (
	"context"
	"fmt"
	"os"
	"testing"

	"g.hz.netease.com/horizon/common"
	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/group/models"
	"gorm.io/gorm"

	"github.com/stretchr/testify/assert"
)

var (
	// use tmp sqlite
	db, _ = orm.NewSqliteDB("")
	ctx   = orm.NewContext(context.TODO(), db)

	group1Id   = uint(1)
	group1Path = "/a"

	notExistID   = uint(100)
	notExistPath = "x"
)

func getGroup1() *models.Group {
	return &models.Group{
		Name:            "1",
		Path:            "a",
		VisibilityLevel: "private",
	}
}

func getGroup2(pid uint) *models.Group {
	return &models.Group{
		Name:            "2",
		Path:            "b",
		VisibilityLevel: "public",
		ParentID:        &pid,
	}
}

func getGroup3(pid uint) *models.Group {
	return &models.Group{
		Name:            "3",
		Path:            "c",
		VisibilityLevel: "public",
		ParentID:        &pid,
	}
}

func init() {
	// create table
	err := db.AutoMigrate(&models.Group{})
	if err != nil {
		fmt.Printf("%+v", err)
		os.Exit(1)
	}
}

func TestCreate(t *testing.T) {
	// normal create
	id, err := Mgr.Create(ctx, getGroup1())
	assert.Nil(t, err)
	assert.Equal(t, id, group1Id)

	// name conflict, parentId: nil
	_, err = Mgr.Create(ctx, getGroup1())
	assert.Equal(t, common.ErrNameConflict, err)

	// normal create, parent: 1
	group2 := getGroup2(id)
	_, err = Mgr.Create(ctx, group2)
	assert.Nil(t, err)
	assert.Equal(t, "/a/b", group2.Path)
	assert.Equal(t, "1 / 2", group2.FullName)

	// name conflict, parentId: 1
	_, err = Mgr.Create(ctx, getGroup2(id))
	assert.Equal(t, common.ErrNameConflict, err)

	// drop table
	db.Where("1 = 1").Delete(&models.Group{})
}

func TestDelete(t *testing.T) {
	id, err := Mgr.Create(ctx, getGroup1())
	assert.Nil(t, err)

	// delete exist record
	err = Mgr.Delete(ctx, id)
	assert.Nil(t, err)

	_, err = Mgr.Get(ctx, id)
	assert.Equal(t, err, gorm.ErrRecordNotFound)

	// delete not exist record
	err = Mgr.Delete(ctx, notExistID)
	assert.Equal(t, err, gorm.ErrRecordNotFound)

	// drop table
	db.Where("1 = 1").Delete(&models.Group{})
}

func TestGet(t *testing.T) {
	id, err := Mgr.Create(ctx, getGroup1())
	assert.Nil(t, err)

	// query exist record
	group1, err := Mgr.Get(ctx, id)
	assert.Nil(t, err)
	assert.NotNil(t, group1.ID)

	// query not exist record
	_, err = Mgr.Get(ctx, notExistID)
	assert.Equal(t, err, gorm.ErrRecordNotFound)

	// drop table
	db.Where("1 = 1").Delete(&models.Group{})
}

func TestGetByPath(t *testing.T) {
	_, err := Mgr.Create(ctx, getGroup1())
	assert.Nil(t, err)

	// query exist record
	group1, err := Mgr.GetByPath(ctx, group1Path)
	assert.Nil(t, err)
	assert.NotNil(t, group1)

	// query not exist record
	_, err = Mgr.GetByPath(ctx /**/, notExistPath)
	assert.Equal(t, err, gorm.ErrRecordNotFound)

	// drop table
	db.Where("1 = 1").Delete(&models.Group{})
}

func TestUpdate(t *testing.T) {
	group1 := getGroup1()
	id, err := Mgr.Create(ctx, group1)
	assert.Nil(t, err)

	// update exist record
	group1.ID = id
	group1.Name = "update1"
	err = Mgr.Update(ctx, group1)
	assert.Nil(t, err)

	// update not exist record
	group1.ID = notExistID
	group1.Name = "update2"
	err = Mgr.Update(ctx, group1)
	assert.Equal(t, err, gorm.ErrRecordNotFound)

	// drop table
	db.Where("1 = 1").Delete(&models.Group{})
}

func TestList(t *testing.T) {
	pid, err := Mgr.Create(ctx, getGroup1())
	assert.Nil(t, err)
	var group2Id, group3Id uint
	group2Id, err = Mgr.Create(ctx, getGroup2(pid))
	assert.Nil(t, err)
	group3Id, err = Mgr.Create(ctx, getGroup3(pid))
	assert.Nil(t, err)

	// page with keywords, items: 1, total: 1
	query := q.New(q.KeyWords{
		"name": "2",
	})
	query.PageNumber = 1
	query.PageSize = 1
	items, total, err := Mgr.List(ctx, query)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(items))
	assert.Equal(t, int64(1), total)

	// page without keywords, items: 1, total: 2
	query.Keywords = q.KeyWords{}
	items, total, err = Mgr.List(ctx, query)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(items))
	assert.Equal(t, int64(3), total)

	// list by parentIdList
	query.Keywords = q.KeyWords{
		"parent_id": []uint{
			pid,
		},
	}
	query.PageSize = 10
	items, _, err = Mgr.List(ctx, query)
	assert.Nil(t, err)
	assert.Equal(t, group2Id, items[0].ID)
	assert.Equal(t, group3Id, items[1].ID)

	// drop table
	db.Where("1 = 1").Delete(&models.Group{})
}

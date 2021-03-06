package myapp

import (
	"fmt"
	"github.com/btnguyen2k/consu/reddo"
	"github.com/btnguyen2k/godal"
	"github.com/btnguyen2k/godal/sql"
	"github.com/btnguyen2k/prom"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"strings"
)

func newSqliteConnection(dir, dbName string) *prom.SqlConnect {
	err := os.MkdirAll(dir, 0711)
	if err != nil {
		panic(err)
	}
	sqlc, err := prom.NewSqlConnect("sqlite3", dir+"/"+dbName+".db", 10000, nil)
	if err != nil {
		panic(err)
	}
	return sqlc
}

func sqliteInitTableGroup(sqlc *prom.SqlConnect, tableName string) {
	sql := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s VARCHAR(64), %s VARCHAR(255), PRIMARY KEY (%s))",
		tableName, sqliteColGroupId, sqliteColGroupName, sqliteColGroupId)
	_, err := sqlc.GetDB().Exec(sql)
	if err != nil {
		panic(err)
	}
}

func sqliteInitTableUser(sqlc *prom.SqlConnect, tableName string) {
	sql := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s VARCHAR(64), %s VARCHAR(64), %s VARCHAR(64), %s VARCHAR(64), PRIMARY KEY (%s))",
		tableName, sqliteColUserUsername, sqliteColUserPassword, sqliteColUserName, sqliteColUserGroupId, sqliteColUserUsername)
	_, err := sqlc.GetDB().Exec(sql)
	if err != nil {
		panic(err)
	}
}

/*----------------------------------------------------------------------*/

func newUserDaoSqlite(sqlc *prom.SqlConnect, tableName string) UserDao {
	dao := &UserDaoSqlite{tableName: tableName}
	dao.GenericDaoSql = sql.NewGenericDaoSql(sqlc, godal.NewAbstractGenericDao(dao))
	dao.SetRowMapper(&sql.GenericRowMapperSql{
		NameTransformation:          sql.NameTransfLowerCase,
		GboFieldToColNameTranslator: map[string]map[string]interface{}{tableName: sqliteMapFieldToColNameUser},
		ColNameToGboFieldTranslator: map[string]map[string]interface{}{tableName: sqliteMapColNameToFieldUser},
		ColumnsListMap:              map[string][]string{tableName: sqliteColsUser},
	})
	return dao
}

const (
	sqliteTableUser       = namespace + "_user"
	sqliteColUserUsername = "uname"
	sqliteColUserPassword = "upwd"
	sqliteColUserName     = "display_name"
	sqliteColUserGroupId  = "gid"
)

var (
	sqliteColsUser              = []string{sqliteColUserUsername, sqliteColUserPassword, sqliteColUserName, sqliteColUserGroupId}
	sqliteMapFieldToColNameUser = map[string]interface{}{fieldUserUsername: sqliteColUserUsername, fieldUserPassword: sqliteColUserPassword, fieldUserName: sqliteColUserName, fieldUserGroupId: sqliteColUserGroupId}
	sqliteMapColNameToFieldUser = map[string]interface{}{sqliteColUserUsername: fieldUserUsername, sqliteColUserPassword: fieldUserPassword, sqliteColUserName: fieldUserName, sqliteColUserGroupId: fieldUserGroupId}
)

type UserDaoSqlite struct {
	*sql.GenericDaoSql
	tableName string
}

// GdaoCreateFilter implements IGenericDao.GdaoCreateFilter
func (dao *UserDaoSqlite) GdaoCreateFilter(_ string, bo godal.IGenericBo) interface{} {
	return map[string]interface{}{sqliteColUserUsername: bo.GboGetAttrUnsafe(fieldUserUsername, reddo.TypeString)}
}

// it is recommended to have a function that transforms godal.IGenericBo to business object and vice versa.
func (dao *UserDaoSqlite) toBo(gbo godal.IGenericBo) *User {
	if gbo == nil {
		return nil
	}
	bo := &User{
		Username: gbo.GboGetAttrUnsafe(fieldUserUsername, reddo.TypeString).(string),
		Password: gbo.GboGetAttrUnsafe(fieldUserPassword, reddo.TypeString).(string),
		Name:     gbo.GboGetAttrUnsafe(fieldUserName, reddo.TypeString).(string),
		GroupId:  gbo.GboGetAttrUnsafe(fieldUserGroupId, reddo.TypeString).(string),
	}
	return bo
}

// it is recommended to have a function that transforms godal.IGenericBo to business object and vice versa.
func (dao *UserDaoSqlite) toGbo(bo *User) godal.IGenericBo {
	if bo == nil {
		return nil
	}
	gbo := godal.NewGenericBo()
	gbo.GboSetAttr(fieldUserUsername, bo.Username)
	gbo.GboSetAttr(fieldUserPassword, bo.Password)
	gbo.GboSetAttr(fieldUserName, bo.Name)
	gbo.GboSetAttr(fieldUserGroupId, bo.GroupId)
	return gbo
}

// Delete implements UserDao.Delete
func (dao *UserDaoSqlite) Delete(bo *User) (bool, error) {
	numRows, err := dao.GdaoDelete(dao.tableName, dao.toGbo(bo))
	return numRows > 0, err
}

// Get implements UserDao.Create
func (dao *UserDaoSqlite) Create(username, encryptedPassword, name, groupId string) (bool, error) {
	bo := &User{
		Username: strings.ToLower(strings.TrimSpace(username)),
		Password: strings.TrimSpace(encryptedPassword),
		Name:     strings.TrimSpace(name),
		GroupId:  strings.ToLower(strings.TrimSpace(groupId)),
	}
	numRows, err := dao.GdaoCreate(dao.tableName, dao.toGbo(bo))
	return numRows > 0, err
}

// Get implements UserDao.Get
func (dao *UserDaoSqlite) Get(username string) (*User, error) {
	gbo, err := dao.GdaoFetchOne(dao.tableName, map[string]interface{}{sqliteColUserUsername: username})
	if err != nil {
		return nil, err
	}
	return dao.toBo(gbo), nil
}

// GetN implements UserDao.GetN
func (dao *UserDaoSqlite) GetN(fromOffset, maxNumRows int) ([]*User, error) {
	gboList, err := dao.GdaoFetchMany(dao.tableName, nil, nil, fromOffset, maxNumRows)
	if err != nil {
		return nil, err
	}
	result := make([]*User, 0)
	for _, gbo := range gboList {
		bo := dao.toBo(gbo)
		result = append(result, bo)
	}
	return result, nil
}

// GetAll implements UserDao.GetAll
func (dao *UserDaoSqlite) GetAll() ([]*User, error) {
	return dao.GetN(0, 0)
}

// Update implements UserDao.Update
func (dao *UserDaoSqlite) Update(bo *User) (bool, error) {
	numRows, err := dao.GdaoUpdate(dao.tableName, dao.toGbo(bo))
	return numRows > 0, err
}

/*----------------------------------------------------------------------*/

func newGroupDaoSqlite(sqlc *prom.SqlConnect, tableName string) GroupDao {
	dao := &GroupDaoSqlite{tableName: tableName}
	dao.GenericDaoSql = sql.NewGenericDaoSql(sqlc, godal.NewAbstractGenericDao(dao))
	dao.SetRowMapper(&sql.GenericRowMapperSql{
		NameTransformation:          sql.NameTransfLowerCase,
		GboFieldToColNameTranslator: map[string]map[string]interface{}{tableName: sqliteMapFieldToColNameGroup},
		ColNameToGboFieldTranslator: map[string]map[string]interface{}{tableName: sqliteMapColNameToFieldGroup},
		ColumnsListMap:              map[string][]string{tableName: sqliteColsGroup},
	})
	return dao
}

const (
	sqliteTableGroup   = namespace + "_group"
	sqliteColGroupId   = "gid"
	sqliteColGroupName = "gname"
)

var (
	sqliteColsGroup              = []string{sqliteColGroupId, sqliteColGroupName}
	sqliteMapFieldToColNameGroup = map[string]interface{}{fieldGroupId: sqliteColGroupId, fieldGroupName: sqliteColGroupName}
	sqliteMapColNameToFieldGroup = map[string]interface{}{sqliteColGroupId: fieldGroupId, sqliteColGroupName: fieldGroupName}
)

type GroupDaoSqlite struct {
	*sql.GenericDaoSql
	tableName string
}

// GdaoCreateFilter implements IGenericDao.GdaoCreateFilter
func (dao *GroupDaoSqlite) GdaoCreateFilter(_ string, bo godal.IGenericBo) interface{} {
	return map[string]interface{}{sqliteColGroupId: bo.GboGetAttrUnsafe(fieldGroupId, reddo.TypeString)}
}

// it is recommended to have a function that transforms godal.IGenericBo to business object and vice versa.
func (dao *GroupDaoSqlite) toBo(gbo godal.IGenericBo) *Group {
	if gbo == nil {
		return nil
	}
	bo := &Group{
		Id:   gbo.GboGetAttrUnsafe(fieldGroupId, reddo.TypeString).(string),
		Name: gbo.GboGetAttrUnsafe(fieldGroupName, reddo.TypeString).(string),
	}
	return bo
}

// it is recommended to have a function that transforms godal.IGenericBo to business object and vice versa.
func (dao *GroupDaoSqlite) toGbo(bo *Group) godal.IGenericBo {
	if bo == nil {
		return nil
	}
	gbo := godal.NewGenericBo()
	gbo.GboSetAttr(fieldGroupId, bo.Id)
	gbo.GboSetAttr(fieldGroupName, bo.Name)
	return gbo
}

// Delete implements GroupDao.Delete
func (dao *GroupDaoSqlite) Delete(bo *Group) (bool, error) {
	numRows, err := dao.GdaoDelete(dao.tableName, dao.toGbo(bo))
	return numRows > 0, err
}

// Get implements GroupDao.Create
func (dao *GroupDaoSqlite) Create(id, name string) (bool, error) {
	bo := &Group{
		Id:   strings.ToLower(strings.TrimSpace(id)),
		Name: strings.TrimSpace(name),
	}
	numRows, err := dao.GdaoCreate(dao.tableName, dao.toGbo(bo))
	return numRows > 0, err
}

// Get implements GroupDao.Get
func (dao *GroupDaoSqlite) Get(id string) (*Group, error) {
	gbo, err := dao.GdaoFetchOne(dao.tableName, map[string]interface{}{sqliteColGroupId: id})
	if err != nil {
		return nil, err
	}
	return dao.toBo(gbo), nil
}

// GetN implements GroupDao.GetN
func (dao *GroupDaoSqlite) GetN(fromOffset, maxNumRows int) ([]*Group, error) {
	gboList, err := dao.GdaoFetchMany(dao.tableName, nil, nil, fromOffset, maxNumRows)
	if err != nil {
		return nil, err
	}
	result := make([]*Group, 0)
	for _, gbo := range gboList {
		bo := dao.toBo(gbo)
		result = append(result, bo)
	}
	return result, nil
}

// GetAll implements GroupDao.GetAll
func (dao *GroupDaoSqlite) GetAll() ([]*Group, error) {
	return dao.GetN(0, 0)
}

// Update implements GroupDao.Update
func (dao *GroupDaoSqlite) Update(bo *Group) (bool, error) {
	numRows, err := dao.GdaoUpdate(dao.tableName, dao.toGbo(bo))
	return numRows > 0, err
}

package user

import (
	"bytes"
	"fmt"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/config"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/module"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/server"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/util"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/wkhttp"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/testutil"
	"github.com/stretchr/testify/assert"
)

func TestAddUser(t *testing.T) {
	s, ctx := testutil.NewTestServer()
	m := NewManager(ctx)
	m.Route(s.GetRoute())
	//清除数据
	err := testutil.CleanAllTables(ctx)
	assert.NoError(t, err)
	w := httptest.NewRecorder()

	req, _ := http.NewRequest("POST", "/v1/manager/adduser", bytes.NewReader([]byte(util.ToJson(map[string]interface{}{
		"name":     "张三",
		"zone":     "0086",
		"phone":    "13600000002",
		"password": "1234567",
	}))))
	req.Header.Set("token", testutil.Token)
	s.GetRoute().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// UID 测试用户ID
var UID = "10000"
var friendUID = "10001"

// Token 测试用户token
var Token = "token122323"

// CleanAllTables 清空所有表
func CleanAllTables(c *config.Context) error {
	var dropSqls []string
	_, err := c.DB().SelectBySql("select  concat('DELETE FROM ','`', table_name,'`') FROM information_schema.tables WHERE table_schema = 'test' and table_name <> 'gorp_migrations'").Load(&dropSqls)
	for _, sql := range dropSqls {
		_, err = c.DB().UpdateBySql(sql).Exec()
		if err != nil {
			return err
		}
	}
	return err
}

// NewTestServer 创建一个测试服务器
func NewTestServer(args ...string) (*server.Server, *config.Context) {
	cfg := config.New()
	cfg.Test = true
	// cfg.TracingOn = true
	// cfg.TracerAddr = "49.235.106.135:6831"
	cfg.DB.MySQLAddr = "root:demo@tcp(127.0.0.1)/test?charset=utf8mb4&parseTime=true"
	cfg.DB.Migration = false
	ctx := config.NewContext(cfg)
	// ctx.Event = event.New(ctx)
	// 先清空旧数据
	err := CleanAllTables(ctx)
	if err != nil {
		panic(err)
	}

	// ctx.Event = event.New(ctx)
	err = ctx.Cache().Set(cfg.Cache.TokenCachePrefix+Token, UID+"@test")
	if err != nil {
		panic(err)
	}

	// _, err = ctx.DB().InsertBySql("insert into `app`(app_id,app_key,status) VALUES('wukongchat',substring(MD5(RAND()),1,20),1)").Exec()
	// if err != nil {
	// 	panic(err)
	// }

	// 创建server
	s := server.New(ctx)
	// ctx.Server = s
	s.GetRoute().UseGin(ctx.Tracer().GinMiddle())
	ctx.SetHttpRoute(s.GetRoute())
	err = module.Setup(ctx)
	if err != nil {
		panic(err)
	}

	return s, ctx

}

func TestLogin(t *testing.T) {
	s, ctx := NewTestServer()
	m := NewManager(ctx)
	m.Route(s.GetRoute())
	//清除数据x
	err := testutil.CleanAllTables(ctx)
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	err = m.userDB.Insert(&Model{
		UID:      "xxx",
		Username: "superAdmin",
		Name:     "超级管理员",
		Password: util.MD5(util.MD5("admiN123456")),
		Role:     string(wkhttp.SuperAdmin),
	})
	assert.NoError(t, err)
	req, _ := http.NewRequest("POST", "/v1/manager/login", bytes.NewReader([]byte(util.ToJson(map[string]interface{}{
		"username": "superAdmin",
		"password": "admiN123456",
	}))))
	req.Header.Set("token", testutil.Token)
	s.GetRoute().ServeHTTP(w, req)
	assert.Equal(t, true, strings.Contains(w.Body.String(), `"uid":"xxx"`))
	assert.Equal(t, true, strings.Contains(w.Body.String(), `"name":"超级管理员"`))
}

func TestBlacklist(t *testing.T) {
	s, ctx := testutil.NewTestServer()
	m := NewManager(ctx)
	m.Route(s.GetRoute())
	//清除数据
	err := testutil.CleanAllTables(ctx)
	assert.NoError(t, err)

	err = m.userDB.Insert(&Model{
		UID:      "xxx",
		Username: "111",
		Name:     "111",
		Password: util.MD5(util.MD5("111")),
	})
	assert.NoError(t, err)
	err = m.userDB.Insert(&Model{
		UID:      "sss",
		Username: "222",
		Name:     "222",
		Password: util.MD5(util.MD5("222")),
	})
	assert.NoError(t, err)
	m.userSettingDB.InsertUserSettingModel(&SettingModel{
		UID:       "xxx",
		ToUID:     "sss",
		Blacklist: 1,
	})
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/manager/blacklist?uid=xxx", nil)
	req.Header.Set("token", testutil.Token)
	s.GetRoute().ServeHTTP(w, req)
}

func TestUpdatePwd(t *testing.T) {
	s, ctx := testutil.NewTestServer()
	m := NewManager(ctx)
	m.Route(s.GetRoute())
	//清除数据
	err := testutil.CleanAllTables(ctx)
	assert.NoError(t, err)

	err = m.userDB.Insert(&Model{
		UID:      testutil.UID,
		Username: "111",
		Name:     "111",
		Password: util.MD5(util.MD5("111")),
	})
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/manager/user/updatepassword", bytes.NewReader([]byte(util.ToJson(map[string]interface{}{
		"new_password": "333333",
		"password":     "111",
	}))))
	req.Header.Set("token", testutil.Token)
	s.GetRoute().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
func TestUserList(t *testing.T) {
	s, ctx := testutil.NewTestServer()
	m := NewManager(ctx)
	// m.Route(s.GetRoute())
	//清除数据
	err := testutil.CleanAllTables(ctx)
	assert.NoError(t, err)
	err = m.userDB.Insert(&Model{
		UID:      util.GenerUUID(),
		ShortNo:  util.GenerUUID(),
		Phone:    "13897655629",
		Username: "111",
		Name:     "111",
		Status:   1,
		GiteeUID: "gitee_uid_1",
		Password: util.MD5(util.MD5("111")),
	})
	assert.NoError(t, err)
	err = m.userDB.Insert(&Model{
		UID:       util.GenerUUID(),
		ShortNo:   util.GenerUUID(),
		Phone:     "13567889876",
		Username:  "222",
		Name:      "222",
		Status:    1,
		GithubUID: "github_uid_1",
		Password:  util.MD5(util.MD5("222")),
	})
	assert.NoError(t, err)
	err = m.userDB.Insert(&Model{
		UID:      util.GenerUUID(),
		ShortNo:  util.GenerUUID(),
		Phone:    "13567987658",
		Username: "333",
		Name:     "333",
		Status:   1,
		WXOpenid: "wx_open_id_1",
		Password: util.MD5(util.MD5("333")),
	})
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/manager/user/list?page_index=1&page_size=10&keyword=222", nil)
	req.Header.Set("token", testutil.Token)
	s.GetRoute().ServeHTTP(w, req)
	assert.Equal(t, true, strings.Contains(w.Body.String(), `"name":"222"`))
}
func TestUserDisablelist(t *testing.T) {
	s, ctx := testutil.NewTestServer()
	m := NewManager(ctx)
	// m.Route(s.GetRoute())
	//清除数据
	err := testutil.CleanAllTables(ctx)
	assert.NoError(t, err)

	err = m.userDB.Insert(&Model{
		UID:      testutil.UID,
		Phone:    "13897655629",
		Username: "111",
		Name:     "111",
		Status:   0,
		Password: util.MD5(util.MD5("111")),
	})
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/manager/user/disablelist?page_index=1&page_size=10", nil)
	req.Header.Set("token", testutil.Token)
	s.GetRoute().ServeHTTP(w, req)
	assert.Equal(t, true, strings.Contains(w.Body.String(), `"name":"111"`))
}
func TestAddAdminUser(t *testing.T) {
	s, ctx := testutil.NewTestServer()
	//清除数据
	err := testutil.CleanAllTables(ctx)
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/manager/user/admin", bytes.NewReader([]byte(util.ToJson(map[string]interface{}{
		"login_name": "admin1",
		"password":   "111",
		"name":       "管理员",
	}))))
	req.Header.Set("token", testutil.Token)
	s.GetRoute().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetAdminUser(t *testing.T) {
	s, ctx := testutil.NewTestServer()
	m := NewManager(ctx)
	//	m.Route(s.GetRoute())
	//清除数据
	err := testutil.CleanAllTables(ctx)
	assert.NoError(t, err)
	err = m.userDB.Insert(&Model{
		UID:      "uid1",
		Name:     "管理员1",
		Role:     "admin",
		Username: "admin",
		ShortNo:  "123",
	})
	assert.NoError(t, err)
	err = m.userDB.Insert(&Model{
		UID:      "uid2",
		Name:     "管理员2",
		Role:     "admin",
		Username: "admin2",
		ShortNo:  "321",
	})
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/manager/user/admin", nil)
	req.Header.Set("token", testutil.Token)
	s.GetRoute().ServeHTTP(w, req)
	// assert.Equal(t, http.StatusOK, w.Code)
	panic(w.Body)
}
func TestDeleteAdminUser(t *testing.T) {
	s, ctx := testutil.NewTestServer()
	m := NewManager(ctx)
	// m.Route(s.GetRoute())
	//清除数据
	err := testutil.CleanAllTables(ctx)
	assert.NoError(t, err)
	uid := "uid1"
	err = m.userDB.Insert(&Model{
		UID:      uid,
		Name:     "管理员1",
		Role:     "admin",
		Username: "admin",
	})
	assert.NoError(t, err)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/v1/manager/user/admin?uid=%s", uid), nil)
	req.Header.Set("token", testutil.Token)
	s.GetRoute().ServeHTTP(w, req)
	// assert.Equal(t, http.StatusOK, w.Code)
	panic(w.Body)
}

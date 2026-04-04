package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/dto/response"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/cmd/server/db_conf"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/compose"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"github.com/subosito/gotenv"
)

type DatabaseInstallReq struct {
	AppKey    string `json:"appKey" validate:"required,oneof=mysql postgresql redis"`
	Port      int    `json:"port" validate:"required"`
	Password  string `json:"password" validate:"required"`
	ImagePath string `json:"imagePath" validate:"required"`
	TaskID    string `json:"taskID"`
}

type DatabaseLinkReq struct {
	AppKey        string `json:"appKey" validate:"required,oneof=mysql postgresql redis"`
	ContainerName string `json:"containerName" validate:"required"`
	Port          int    `json:"port" validate:"required"`
	Password      string `json:"password" validate:"required"`
	Username      string `json:"username"`
}

type IDatabaseInstallService interface {
	CheckInstalled(appKey string) (*response.AppInstalledCheck, error)
	Install(req DatabaseInstallReq) error
	ListContainers(appKey string) ([]OpenrestyContainerInfo, error)
	LinkContainer(req DatabaseLinkReq) error
}

type DatabaseInstallService struct{}

func NewIDatabaseInstallService() IDatabaseInstallService {
	return &DatabaseInstallService{}
}

var dbAppMeta = map[string]struct {
	Name         string
	ShortDescZh  string
	ShortDescEn  string
	DefaultPort  int
	ContainerPfx string
	Version      string
	Image        string
	Compose      func() []byte
	Keywords     []string
}{
	constant.AppMysql: {
		Name:         "MySQL",
		ShortDescZh:  "MySQL 数据库",
		ShortDescEn:  "MySQL Database",
		DefaultPort:  3306,
		ContainerPfx: "1Panel-mysql-1",
		Version:      "8.0",
		Image:        "mysql",
		Compose:      func() []byte { return db_conf.MysqlDockerCompose },
		Keywords:     []string{"mysql", "mariadb"},
	},
	constant.AppPostgresql: {
		Name:         "PostgreSQL",
		ShortDescZh:  "PostgreSQL 数据库",
		ShortDescEn:  "PostgreSQL Database",
		DefaultPort:  5432,
		ContainerPfx: "1Panel-postgresql-1",
		Version:      "16",
		Image:        "postgres",
		Compose:      func() []byte { return db_conf.PostgresqlDockerCompose },
		Keywords:     []string{"postgres", "postgresql"},
	},
	constant.AppRedis: {
		Name:         "Redis",
		ShortDescZh:  "Redis 缓存数据库",
		ShortDescEn:  "Redis Cache Database",
		DefaultPort:  6379,
		ContainerPfx: "1Panel-redis-1",
		Version:      "7",
		Image:        "redis",
		Compose:      func() []byte { return db_conf.RedisDockerCompose },
		Keywords:     []string{"redis"},
	},
}

func (d *DatabaseInstallService) CheckInstalled(appKey string) (*response.AppInstalledCheck, error) {
	installService := &AppInstallService{}
	return installService.CheckExist(request.AppInstalledInfo{Key: appKey})
}

func (d *DatabaseInstallService) Install(req DatabaseInstallReq) error {
	meta, ok := dbAppMeta[req.AppKey]
	if !ok {
		return fmt.Errorf("unsupported database app: %s", req.AppKey)
	}

	if err := docker.CreateDefaultDockerNetwork(); err != nil {
		return buserr.WithDetail("ErrCreateNetwork", err.Error(), err)
	}

	existing, _ := d.CheckInstalled(req.AppKey)
	if existing != nil && existing.IsExist {
		return buserr.New("ErrAppIsExist")
	}

	app, err := ensureDBApp(req.AppKey, meta)
	if err != nil {
		return err
	}
	appDetail, err := ensureDBAppDetail(app.ID, req.AppKey, meta)
	if err != nil {
		return err
	}

	containerName := meta.ContainerPfx
	if exist, _ := appInstallRepo.GetFirst(appInstallRepo.WithContainerName(containerName)); exist.ID > 0 {
		return buserr.New("ErrContainerName")
	}

	passwordKey := "PANEL_DB_ROOT_PASSWORD"
	if req.AppKey == constant.AppRedis {
		passwordKey = "PANEL_REDIS_ROOT_PASSWORD"
	}

	env := map[string]string{
		"CONTAINER_NAME":      containerName,
		"PANEL_APP_PORT_HTTP": strconv.Itoa(req.Port),
		passwordKey:           req.Password,
	}

	envBytes, _ := json.Marshal(env)
	composeContent := string(meta.Compose())

	appInstall := &model.AppInstall{
		Name:          req.AppKey,
		AppId:         app.ID,
		AppDetailId:   appDetail.ID,
		Version:       meta.Version,
		Status:        constant.StatusInstalling,
		Env:           string(envBytes),
		HttpPort:      req.Port,
		DockerCompose: composeContent,
		ContainerName: containerName,
		ServiceName:   containerName,
		App:           *app,
	}

	installDir := path.Join(global.Dir.AppInstallDir, req.AppKey, req.AppKey)
	fileOp := files.NewFileOp()
	if fileOp.Stat(installDir) {
		_ = fileOp.DeleteDir(installDir)
	}

	installTask, err := task.NewTaskWithOps(req.AppKey, task.TaskInstall, task.TaskScopeApp, req.TaskID, 0)
	if err != nil {
		return err
	}

	installTask.AddSubTaskWithOps("prepare", func(t *task.Task) error {
		if err := fileOp.CreateDir(installDir, constant.DirPerm); err != nil {
			return err
		}
		for _, sub := range []string{"data", "conf", "log"} {
			_ = fileOp.CreateDir(path.Join(installDir, sub), constant.DirPerm)
		}

		envPath := path.Join(installDir, ".env")
		_ = gotenv.Write(env, envPath)

		composePath := path.Join(installDir, "docker-compose.yml")
		if err := fileOp.SaveFileWithByte(composePath, []byte(composeContent), constant.DirPerm); err != nil {
			return err
		}

		if err := appInstallRepo.Create(context.Background(), appInstall); err != nil {
			return err
		}
		return nil
	}, func(t *task.Task) {
		if appInstall != nil && appInstall.ID > 0 {
			_ = appInstallRepo.Delete(context.Background(), *appInstall)
		}
		_ = fileOp.DeleteDir(installDir)
	}, 0, 0)

	installTask.AddSubTaskWithOps("load-image", func(t *task.Task) error {
		file, err := os.Open(req.ImagePath)
		if err != nil {
			return fmt.Errorf("failed to open image tar: %w", err)
		}
		defer file.Close()
		dockerCli, err := docker.NewDockerClient()
		if err != nil {
			return err
		}
		defer dockerCli.Close()
		res, err := dockerCli.ImageLoad(context.TODO(), file)
		if err != nil {
			return fmt.Errorf("failed to load docker image: %w", err)
		}
		defer res.Body.Close()
		return nil
	}, nil, 0, 0)

	installTask.AddSubTaskWithOps("start", func(t *task.Task) error {
		appInstall.Status = constant.StatusRunning
		upOut, err := compose.Up(appInstall.GetComposePath())
		if err != nil {
			appInstall.Status = constant.StatusError
			appInstall.Message = upOut
			_ = appInstallRepo.Save(context.Background(), appInstall)
			return err
		}
		_ = appInstallRepo.Save(context.Background(), appInstall)
		return nil
	}, func(t *task.Task) {
		if appInstall != nil && appInstall.ID > 0 {
			_ = appInstallRepo.Delete(context.Background(), *appInstall)
		}
		_ = fileOp.DeleteDir(installDir)
	}, 0, 0)

	installTask.AddSubTaskWithOps("create-db-record", func(t *task.Task) error {
		params := map[string]interface{}{
			passwordKey: req.Password,
		}
		return CreateDB(context.Background(), *app, appInstall, params)
	}, nil, 0, 0)

	go func() {
		_ = installTask.Execute()
	}()

	return nil
}

func (d *DatabaseInstallService) ListContainers(appKey string) ([]OpenrestyContainerInfo, error) {
	meta, ok := dbAppMeta[appKey]
	if !ok {
		return nil, fmt.Errorf("unsupported database app: %s", appKey)
	}

	cli, err := docker.NewClient()
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	containers, err := cli.ListAllContainers()
	if err != nil {
		return nil, err
	}

	var result []OpenrestyContainerInfo
	for _, c := range containers {
		name := ""
		if len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}
		img := strings.ToLower(c.Image)
		isValid := false
		for _, kw := range meta.Keywords {
			if strings.Contains(img, kw) {
				isValid = true
				break
			}
		}
		result = append(result, OpenrestyContainerInfo{
			Name:    name,
			Image:   c.Image,
			Status:  c.State,
			IsValid: isValid,
		})
	}
	return result, nil
}

func (d *DatabaseInstallService) LinkContainer(req DatabaseLinkReq) error {
	meta, ok := dbAppMeta[req.AppKey]
	if !ok {
		return fmt.Errorf("unsupported database app: %s", req.AppKey)
	}

	cli, err := docker.NewDockerClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	inspect, err := cli.ContainerInspect(context.Background(), req.ContainerName)
	if err != nil {
		return fmt.Errorf("container %s not found: %w", req.ContainerName, err)
	}

	existing, _ := d.CheckInstalled(req.AppKey)
	if existing != nil && existing.IsExist {
		return buserr.New("ErrAppIsExist")
	}

	app, err := ensureDBApp(req.AppKey, meta)
	if err != nil {
		return err
	}
	appDetail, err := ensureDBAppDetail(app.ID, req.AppKey, meta)
	if err != nil {
		return err
	}

	version := inspect.Config.Image
	if parts := strings.SplitN(version, ":", 2); len(parts) == 2 {
		version = parts[1]
	}

	passwordKey := "PANEL_DB_ROOT_PASSWORD"
	if req.AppKey == constant.AppRedis {
		passwordKey = "PANEL_REDIS_ROOT_PASSWORD"
	}

	env := map[string]string{
		"CONTAINER_NAME":      req.ContainerName,
		"PANEL_APP_PORT_HTTP": strconv.Itoa(req.Port),
		passwordKey:           req.Password,
	}
	envBytes, _ := json.Marshal(env)

	username := req.Username
	if username == "" && (req.AppKey == constant.AppMysql) {
		username = "root"
	}
	if username == "" && req.AppKey == constant.AppPostgresql {
		username = "postgres"
	}

	appInstall := &model.AppInstall{
		Name:          req.AppKey,
		AppId:         app.ID,
		AppDetailId:   appDetail.ID,
		Version:       version,
		Status:        constant.StatusRunning,
		Env:           string(envBytes),
		HttpPort:      req.Port,
		DockerCompose: "",
		ContainerName: req.ContainerName,
		ServiceName:   req.ContainerName,
		App:           *app,
	}

	if err := appInstallRepo.Create(context.Background(), appInstall); err != nil {
		return err
	}

	params := map[string]interface{}{
		passwordKey: req.Password,
	}
	if username != "" {
		params["PANEL_DB_ROOT_USER"] = username
	}
	_ = CreateDB(context.Background(), *app, appInstall, params)

	return nil
}

func ensureDBApp(appKey string, meta struct {
	Name         string
	ShortDescZh  string
	ShortDescEn  string
	DefaultPort  int
	ContainerPfx string
	Version      string
	Image        string
	Compose      func() []byte
	Keywords     []string
}) (*model.App, error) {
	app, err := appRepo.GetFirst(appRepo.WithKey(appKey))
	if err == nil {
		return &app, nil
	}

	newApp := &model.App{
		Name:        meta.Name,
		Key:         appKey,
		ShortDescZh: meta.ShortDescZh,
		ShortDescEn: meta.ShortDescEn,
		Type:        "runtime",
		Status:      constant.AppNormal,
		Limit:       0,
		Resource:    constant.AppResourceLocal,
	}
	if err := appRepo.Create(context.Background(), newApp); err != nil {
		return nil, fmt.Errorf("failed to create %s app record: %w", appKey, err)
	}
	return newApp, nil
}

func ensureDBAppDetail(appID uint, appKey string, meta struct {
	Name         string
	ShortDescZh  string
	ShortDescEn  string
	DefaultPort  int
	ContainerPfx string
	Version      string
	Image        string
	Compose      func() []byte
	Keywords     []string
}) (*model.AppDetail, error) {
	detail, err := appDetailRepo.GetFirst(appDetailRepo.WithAppId(appID))
	if err == nil {
		return &detail, nil
	}

	newDetail := model.AppDetail{
		AppId:         appID,
		Version:       meta.Version,
		DockerCompose: string(meta.Compose()),
		Status:        constant.AppNormal,
	}
	if err := appDetailRepo.BatchCreate(context.Background(), []model.AppDetail{newDetail}); err != nil {
		return nil, fmt.Errorf("failed to create %s app detail: %w", appKey, err)
	}
	created, err := appDetailRepo.GetFirst(appDetailRepo.WithAppId(appID))
	if err != nil {
		return nil, err
	}
	return &created, nil
}

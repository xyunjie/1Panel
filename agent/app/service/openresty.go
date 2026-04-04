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
	"github.com/1Panel-dev/1Panel/agent/cmd/server/nginx_conf"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/compose"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"github.com/subosito/gotenv"
)

const (
	defaultOpenrestyVersion = "1.27.1.2-3-3-focal"
	defaultOpenrestyName    = "openresty"
	defaultContainerName    = "1Panel-openresty-1"
)

type OpenrestyInstallReq struct {
	HttpPort  int    `json:"httpPort" validate:"required"`
	HttpsPort int    `json:"httpsPort" validate:"required"`
	ImagePath string `json:"imagePath" validate:"required"`
	TaskID    string `json:"taskID"`
}

type OpenrestyLinkReq struct {
	ContainerName string `json:"containerName" validate:"required"`
	HttpPort      int    `json:"httpPort" validate:"required"`
	HttpsPort     int    `json:"httpsPort" validate:"required"`
}

type OpenrestyContainerInfo struct {
	Name    string `json:"name"`
	Image   string `json:"image"`
	Status  string `json:"status"`
	IsValid bool   `json:"isValid"`
}

type IOpenrestyService interface {
	CheckInstalled() (*response.AppInstalledCheck, error)
	Install(req OpenrestyInstallReq) error
	ListContainers() ([]OpenrestyContainerInfo, error)
	LinkContainer(req OpenrestyLinkReq) error
}

type OpenrestyService struct{}

func NewIOpenrestyService() IOpenrestyService {
	return &OpenrestyService{}
}

func (o *OpenrestyService) CheckInstalled() (*response.AppInstalledCheck, error) {
	installService := &AppInstallService{}
	return installService.CheckExist(request.AppInstalledInfo{Key: constant.AppOpenresty})
}

func (o *OpenrestyService) Install(req OpenrestyInstallReq) error {
	if err := docker.CreateDefaultDockerNetwork(); err != nil {
		return buserr.WithDetail("ErrCreateNetwork", err.Error(), err)
	}

	existing, _ := o.CheckInstalled()
	if existing != nil && existing.IsExist {
		return buserr.New("ErrAppIsExist")
	}

	app, err := ensureOpenrestyApp()
	if err != nil {
		return err
	}
	appDetail, err := ensureOpenrestyAppDetail(app.ID)
	if err != nil {
		return err
	}

	containerName := defaultContainerName
	if exist, _ := appInstallRepo.GetFirst(appInstallRepo.WithContainerName(containerName)); exist.ID > 0 {
		return buserr.New("ErrContainerName")
	}

	websiteDir := path.Join(global.Dir.DataDir, "www")

	env := map[string]string{
		"CONTAINER_NAME":              containerName,
		"PANEL_APP_PORT_HTTP":         strconv.Itoa(req.HttpPort),
		"PANEL_APP_PORT_HTTPS":        strconv.Itoa(req.HttpsPort),
		"CONTAINER_PACKAGE_URL":       "http://archive.ubuntu.com/ubuntu/",
		"RESTY_CONFIG_OPTIONS_MORE":   "",
		"RESTY_ADD_PACKAGE_BUILDDEPS": "",
		"WEBSITE_DIR":                 websiteDir,
	}

	envBytes, _ := json.Marshal(env)
	composeContent := string(nginx_conf.OpenrestyDockerCompose)

	appInstall := &model.AppInstall{
		Name:          defaultOpenrestyName,
		AppId:         app.ID,
		AppDetailId:   appDetail.ID,
		Version:       defaultOpenrestyVersion,
		Status:        constant.StatusInstalling,
		Env:           string(envBytes),
		HttpPort:      req.HttpPort,
		HttpsPort:     req.HttpsPort,
		DockerCompose: composeContent,
		ContainerName: containerName,
		ServiceName:   containerName,
		App:           *app,
	}

	installDir := path.Join(global.Dir.AppInstallDir, constant.AppOpenresty, defaultOpenrestyName)
	fileOp := files.NewFileOp()
	if fileOp.Stat(installDir) {
		_ = fileOp.DeleteDir(installDir)
	}

	installTask, err := task.NewTaskWithOps(defaultOpenrestyName, task.TaskInstall, task.TaskScopeApp, req.TaskID, 0)
	if err != nil {
		return err
	}

	installTask.AddSubTaskWithOps("prepare", func(t *task.Task) error {
		if err := fileOp.CreateDir(installDir, constant.DirPerm); err != nil {
			return err
		}
		confDir := path.Join(installDir, "conf")
		if err := fileOp.CreateDir(confDir, constant.DirPerm); err != nil {
			return err
		}
		for _, sub := range []string{"log", "build", "1pwaf"} {
			if err := fileOp.CreateDir(path.Join(installDir, sub), constant.DirPerm); err != nil {
				return err
			}
		}

		if err := fileOp.SaveFileWithByte(path.Join(confDir, "nginx.conf"), nginx_conf.OpenrestyNginxConf, constant.DirPerm); err != nil {
			return err
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

	installTask.AddSubTaskWithOps("init", func(t *task.Task) error {
		return initOpenrestyWebsiteDirs(websiteDir)
	}, nil, 0, 0)

	go func() {
		_ = installTask.Execute()
	}()

	return nil
}

func (o *OpenrestyService) ListContainers() ([]OpenrestyContainerInfo, error) {
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
		isValid := strings.Contains(img, "openresty") || strings.Contains(img, "nginx")
		result = append(result, OpenrestyContainerInfo{
			Name:    name,
			Image:   c.Image,
			Status:  c.State,
			IsValid: isValid,
		})
	}
	return result, nil
}

func (o *OpenrestyService) LinkContainer(req OpenrestyLinkReq) error {
	cli, err := docker.NewDockerClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	inspect, err := cli.ContainerInspect(context.Background(), req.ContainerName)
	if err != nil {
		return fmt.Errorf("container %s not found: %w", req.ContainerName, err)
	}

	img := strings.ToLower(inspect.Config.Image)
	if !strings.Contains(img, "openresty") && !strings.Contains(img, "nginx") {
		return buserr.New("ErrNotOpenrestyContainer")
	}

	existing, _ := o.CheckInstalled()
	if existing != nil && existing.IsExist {
		return buserr.New("ErrAppIsExist")
	}

	app, err := ensureOpenrestyApp()
	if err != nil {
		return err
	}
	appDetail, err := ensureOpenrestyAppDetail(app.ID)
	if err != nil {
		return err
	}

	version := inspect.Config.Image
	if parts := strings.SplitN(version, ":", 2); len(parts) == 2 {
		version = parts[1]
	}

	websiteDir := path.Join(global.Dir.DataDir, "www")
	env := map[string]string{
		"CONTAINER_NAME":       req.ContainerName,
		"PANEL_APP_PORT_HTTP":  strconv.Itoa(req.HttpPort),
		"PANEL_APP_PORT_HTTPS": strconv.Itoa(req.HttpsPort),
		"WEBSITE_DIR":          websiteDir,
	}
	envBytes, _ := json.Marshal(env)

	appInstall := &model.AppInstall{
		Name:          defaultOpenrestyName,
		AppId:         app.ID,
		AppDetailId:   appDetail.ID,
		Version:       version,
		Status:        constant.StatusRunning,
		Env:           string(envBytes),
		HttpPort:      req.HttpPort,
		HttpsPort:     req.HttpsPort,
		DockerCompose: "",
		ContainerName: req.ContainerName,
		ServiceName:   req.ContainerName,
		App:           *app,
	}

	if err := appInstallRepo.Create(context.Background(), appInstall); err != nil {
		return err
	}

	_ = initOpenrestyWebsiteDirs(websiteDir)

	return nil
}

func initOpenrestyWebsiteDirs(websiteDir string) error {
	fileOp := files.NewFileOp()
	for _, sub := range []string{"conf.d", "stream.d", "sites", "1pwaf"} {
		d := path.Join(websiteDir, sub)
		if !fileOp.Stat(d) {
			if err := fileOp.CreateDir(d, constant.DirPerm); err != nil {
				return err
			}
		}
	}
	return nil
}

func ensureOpenrestyApp() (*model.App, error) {
	app, err := appRepo.GetFirst(appRepo.WithKey(constant.AppOpenresty))
	if err == nil {
		return &app, nil
	}

	newApp := &model.App{
		Name:        "OpenResty",
		Key:         constant.AppOpenresty,
		ShortDescZh: "OpenResty Web 服务器",
		ShortDescEn: "OpenResty Web Server",
		Type:        "website",
		Status:      constant.AppNormal,
		Limit:       1,
		Resource:    constant.AppResourceLocal,
	}
	if err := appRepo.Create(context.Background(), newApp); err != nil {
		return nil, fmt.Errorf("failed to create openresty app record: %w", err)
	}
	return newApp, nil
}

func ensureOpenrestyAppDetail(appID uint) (*model.AppDetail, error) {
	detail, err := appDetailRepo.GetFirst(appDetailRepo.WithAppId(appID))
	if err == nil {
		return &detail, nil
	}

	newDetail := model.AppDetail{
		AppId:         appID,
		Version:       defaultOpenrestyVersion,
		DockerCompose: string(nginx_conf.OpenrestyDockerCompose),
		Status:        constant.AppNormal,
	}
	if err := appDetailRepo.BatchCreate(context.Background(), []model.AppDetail{newDetail}); err != nil {
		return nil, fmt.Errorf("failed to create openresty app detail: %w", err)
	}
	created, err := appDetailRepo.GetFirst(appDetailRepo.WithAppId(appID))
	if err != nil {
		return nil, err
	}
	return &created, nil
}

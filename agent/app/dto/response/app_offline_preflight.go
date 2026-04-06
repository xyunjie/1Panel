package response

type AppOfflineInstallPreflight struct {
	Ready             bool   `json:"ready"`
	AppKey            string `json:"appKey"`
	RequiredImage     string `json:"requiredImage"`
	DockerReady       bool   `json:"dockerReady"`
	DockerMessage     string `json:"dockerMessage"`
	ImageReady        bool   `json:"imageReady"`
	ImageMessage      string `json:"imageMessage"`
	InstallDir        string `json:"installDir"`
	InstallDirReady   bool   `json:"installDirReady"`
	InstallDirMessage string `json:"installDirMessage"`
	WebsiteDir        string `json:"websiteDir"`
	WebsiteDirReady   bool   `json:"websiteDirReady"`
	WebsiteDirMessage string `json:"websiteDirMessage"`
	DiskReady         bool   `json:"diskReady"`
	DiskMessage       string `json:"diskMessage"`
	DiskUsedPercent   int    `json:"diskUsedPercent"`
	DiskAvailable     uint64 `json:"diskAvailable"`
	DiskWarning       int    `json:"diskWarning"`
}

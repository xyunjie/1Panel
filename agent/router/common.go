package router

func commonGroups() []CommonRouter {
	return []CommonRouter{
		&DashboardRouter{},
		&HostRouter{},
		&ContainerRouter{},
		&LogRouter{},
		&FileRouter{},
		&ToolboxRouter{},
		&CronjobRouter{},
		&BackupRouter{},
		&SettingRouter{},
		&WebsiteRouter{},
		&WebsiteDnsAccountRouter{},
		&WebsiteAcmeAccountRouter{},
		&WebsiteSSLRouter{},
		&DatabaseRouter{},
		&NginxRouter{},
		&RuntimeRouter{},
		&ProcessRouter{},
		&WebsiteCARouter{},
		&AIToolsRouter{},
		&GroupRouter{},
		&AlertRouter{},
	}
}

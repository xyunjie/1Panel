package job

type backup struct{}

func NewBackupJob() *backup {
	return &backup{}
}

func (b *backup) Run() {
	// Token refresh for removed cloud providers (OneDrive/ALIYUN) is no longer needed
}

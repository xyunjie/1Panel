package service

import (
	"github.com/1Panel-dev/1Panel/agent/app/dto"
)

type AlertSender struct {
	alert     dto.AlertDTO
	quotaType string
}

func NewAlertSender(alert dto.AlertDTO, quotaType string) *AlertSender {
	return &AlertSender{
		alert:     alert,
		quotaType: quotaType,
	}
}

// Send is a no-op in the intranet edition (external notification channels removed).
func (s *AlertSender) Send(quota string, params []dto.Param) {}

// ResourceSend is a no-op in the intranet edition (external notification channels removed).
func (s *AlertSender) ResourceSend(quota string, params []dto.Param) {}

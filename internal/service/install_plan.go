package service

import "dst-server-ctl/internal/domain"

type InstallPlanner struct{}

func NewInstallPlanner() *InstallPlanner {
	return &InstallPlanner{}
}

func (p *InstallPlanner) Plan(state domain.InstallationState) domain.InstallPlan {
	var steps []domain.InstallStep

	if state.SteamCMDInstalledAt == nil {
		steps = append(steps, domain.InstallStep{
			Type:        domain.TaskTypeInstallSteamCMD,
			Description: "Install SteamCMD into the managed root",
		})
	}
	if state.DSTInstalledAt == nil {
		steps = append(steps, domain.InstallStep{
			Type:        domain.TaskTypeInstallDST,
			Description: "Install Don't Starve Together dedicated server app 343050",
		})
	}

	return domain.InstallPlan{Steps: steps}
}

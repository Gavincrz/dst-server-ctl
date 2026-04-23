package steamcmd

import "dst-server-ctl/internal/domain"

const dstDedicatedServerAppID = "343050"

type CommandPlan struct {
	Name string
	Args []string
}

func InstallDSTPlan(layout domain.ManagedLayout) CommandPlan {
	return CommandPlan{
		Name: "steamcmd",
		Args: []string{
			"+force_install_dir", layout.DST,
			"+login", "anonymous",
			"+app_update", dstDedicatedServerAppID, "validate",
			"+quit",
		},
	}
}

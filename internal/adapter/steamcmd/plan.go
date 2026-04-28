package steamcmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"dst-server-ctl/internal/domain"
)

const dstDedicatedServerAppID = "343050"

type CommandPlan struct {
	Name string
	Args []string
}

func InstallDSTPlan(layout domain.ManagedLayout) CommandPlan {
	return CommandPlan{
		Name: filepath.Join(layout.SteamCMD, "steamcmd.sh"),
		Args: []string{
			"+force_install_dir", layout.DST,
			"+login", "anonymous",
			"+app_update", dstDedicatedServerAppID, "validate",
			"+quit",
		},
	}
}

func RemoteVersionPlan(layout domain.ManagedLayout) CommandPlan {
	return CommandPlan{
		Name: filepath.Join(layout.SteamCMD, "steamcmd.sh"),
		Args: []string{
			"+login", "anonymous",
			"+app_info_update", "1",
			"+app_info_print", dstDedicatedServerAppID,
			"+quit",
		},
	}
}

func LocalManifestPath(layout domain.ManagedLayout) string {
	return filepath.Join(layout.SteamCMD, "steamapps", "appmanifest_"+dstDedicatedServerAppID+".acf")
}

var localBuildIDPattern = regexp.MustCompile(`"buildid"\s+"([^"]+)"`)
var remotePublicBuildIDPattern = regexp.MustCompile(`(?s)"branches"\s*\{.*?"public"\s*\{.*?"buildid"\s*"([^"]+)"`)

func ParseLocalVersion(content string) (string, error) {
	matches := localBuildIDPattern.FindStringSubmatch(content)
	if len(matches) != 2 {
		return "", fmt.Errorf("parse local build id: buildid not found")
	}
	return matches[1], nil
}

func ParseRemoteVersion(output string) (string, error) {
	matches := remotePublicBuildIDPattern.FindStringSubmatch(output)
	if len(matches) == 2 {
		return matches[1], nil
	}

	matches = localBuildIDPattern.FindStringSubmatch(output)
	if len(matches) == 2 {
		return matches[1], nil
	}

	return "", fmt.Errorf("parse remote build id: buildid not found")
}

func ReadLocalVersion(layout domain.ManagedLayout) (string, error) {
	content, err := os.ReadFile(LocalManifestPath(layout))
	if err != nil {
		return "", err
	}
	return ParseLocalVersion(string(content))
}

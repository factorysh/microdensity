package run

import (
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"

	"github.com/docker/cli/cli/config/configfile"
	dtypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
)

func dockerConfig() (*configfile.ConfigFile, error) {
	dockercfg := &configfile.ConfigFile{}
	var home string
	me, err := user.Current()
	if err != nil {
		// In docker container, you can -u an unknown user
		home = os.TempDir()
	} else {
		home = me.HomeDir
	}

	pth := filepath.Clean(path.Join(home, "/.docker/config.json"))
	if !strings.HasPrefix(pth, home) {
		panic("no path escape: " + pth)
	}
	_, err = os.Stat(pth)
	if err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(path.Join(home, "/.docker"), 0700)
			if err != nil {
				return nil, err
			}
			dockercfg = configfile.New(pth)
		} else {
			return nil, err
		}
	} else {
		f, err := os.Open(pth)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		err = dockercfg.LoadFromReader(f)
		if err != nil {
			return nil, err
		}
	}
	return dockercfg, nil
}

// Lazy network creation
func ensureNetwork(cli *client.Client, networkName string) error {
	networks, err := cli.NetworkList(context.Background(), dtypes.NetworkListOptions{
		Filters: filters.NewArgs(filters.KeyValuePair{
			Key:   "name",
			Value: networkName,
		},
		)})

	if err != nil {
		return err
	}

	if len(networks) == 0 {
		_, err = cli.NetworkCreate(context.Background(), networkName, dtypes.NetworkCreate{})
		if err != nil {
			return err

		}
	}

	return err
}

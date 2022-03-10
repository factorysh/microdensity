package task

/*
Users push task.Task to a queue.Queue
*/

import (
	"context"
	"errors"
	"fmt"
	"io"
	"regexp"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
)

type State int

var (
	sha  = regexp.MustCompile(`^[0-9a-f]+$`)
	name = regexp.MustCompile(`^[0-9a-zA-Z\-%_]+$`)
)

const (
	Ready State = iota
	Running
	Canceled
	Failed
	Done
	Interrupted
)

func (s State) String() string {
	return []string{"Ready", "Running", "Canceled", "Failed", "Done", "Interrupted"}[s]
}

type Task struct {
	Id uuid.UUID `json:"id"`
	// Run is used to save the name of the main/master container for this service
	Run      string                 `json:"run"`
	Service  string                 `json:"service"`
	Project  string                 `json:"project"`
	Branch   string                 `json:"branch"`
	Commit   string                 `json:"commit"`
	Creation time.Time              `json:"creation"`
	Args     map[string]interface{} `json:"Args"`
	State    State
}

func (t *Task) Validate() error {
	if t.Id == uuid.Nil {
		return errors.New("empty id not allowed")
	}
	if len(name.FindIndex([]byte(t.Project))) != 2 {
		return fmt.Errorf("project name must be url escaped, without any strange letter : %s", t.Project)
	}
	if len(name.FindIndex([]byte(t.Branch))) != 2 {
		return fmt.Errorf("branch name must be url escaped, without any strange letter : %s", t.Branch)
	}
	if len(sha.FindIndex([]byte(t.Commit))) != 2 {
		return fmt.Errorf("bad commit format : %s", t.Commit)
	}
	return nil
}

// Logs steam logs of the current run
func (t *Task) Logs(ctx context.Context, follow bool) (io.ReadCloser, error) {
	mainName := fmt.Sprintf("%s_%s_%v", t.Service, t.Run, t.Id)
	docker, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}

	return docker.ContainerLogs(ctx, mainName, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: true,
		Follow:     follow,
	})
}

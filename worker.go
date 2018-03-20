package main

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	_ "github.com/docker/distribution"
	docker "github.com/docker/docker/client"
)

// Work performs the work denoted by j and returns the result path upon
// successful completion.
//
// TODO: introduce build type
// TODO: log fs command outputs
func Work(ctx context.Context, j *Job, fs FileSystem) (string, error) {
	// TODO: this is racy; do this using a map and a mutex
	_, err := os.Stat(j.PendingBuildPath)
	// TODO: log repetitions
	if err == nil {
		t := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-ctx.Done():
				return "", workErr("context cancelled while waiting for pending build", nil)
			case <-t.C:
				_, err = os.Stat(j.ReadyBuildPath)
				if err == nil {
					return j.ReadyBuildPath, nil
				} else if !os.IsNotExist(err) {
					return "", workErr("could not wait for ready build", err)
				}
			}
		}
	} else if !os.IsNotExist(err) {
		return "", workErr("could not check for pending build", err)
	}

	_, err = os.Stat(j.ReadyBuildPath)
	if err == nil {
		return j.ReadyBuildPath, nil
	} else if !os.IsNotExist(err) {
		return "", workErr("could not check for ready path", err)
	}

	err = BootstrapProject(j)
	if err != nil {
		return "", workErr("could not bootstrap project", err)
	}

	src, err := filepath.EvalSymlinks(j.LatestBuildPath)
	if err == nil {
		if j.Group != "" {
			err = fs.Clone(src, j.PendingBuildPath)
			if err != nil {
				return "", workErr("could not clone latest build result", err)
			}
			err = os.RemoveAll(filepath.Join(j.PendingBuildPath, DataDir, ParamsDir))
			if err != nil {
				return "", workErr("could not remove params dir", err)
			}
			err = EnsureDirExists(filepath.Join(j.PendingBuildPath, DataDir, ParamsDir))
			if err != nil {
				return "", workErr("could not ensure directory exists", err)
			}
		}
	} else if os.IsNotExist(err) {
		err = fs.Create(j.PendingBuildPath)
		if err != nil {
			return "", workErr("could not create pending build path", err)
		}
		err = EnsureDirExists(filepath.Join(j.PendingBuildPath, DataDir))
		if err != nil {
			return "", workErr("could not ensure directory exists", err)
		}
		err = EnsureDirExists(filepath.Join(j.PendingBuildPath, DataDir, CacheDir))
		if err != nil {
			return "", workErr("could not ensure directory exists", err)
		}
		err = EnsureDirExists(filepath.Join(j.PendingBuildPath, DataDir, ArtifactsDir))
		if err != nil {
			return "", workErr("could not ensure directory exists", err)
		}
		err = EnsureDirExists(filepath.Join(j.PendingBuildPath, DataDir, ParamsDir))
		if err != nil {
			return "", workErr("could not ensure directory exists", err)
		}
	} else {
		return "", workErr("could not read latest build link", err)
	}

	for k, v := range j.Params {
		err = ioutil.WriteFile(filepath.Join(j.PendingBuildPath, DataDir, ParamsDir, k), []byte(v), 0644)
		if err != nil {
			return "", workErr("could not write param file", err)
		}
	}

	out, err := os.Create(j.BuildLogPath)
	if err != nil {
		return "", workErr("could not create build log file", err)
	}

	// TODO: we should check the error here. However, it's not so simple
	// cause we must always close the file even if eg. BuildImage() failed
	defer out.Close()

	client, err := docker.NewEnvClient()
	if err != nil {
		return "", workErr("could not create docker client", err)
	}

	err = j.BuildImage(ctx, client, out)
	if err != nil {
		return "", workErr("could not build docker image", err)
	}

	err = j.StartContainer(ctx, client, out)
	if err != nil {
		return "", workErr("could not start docker container", err)
	}

	err = os.Rename(j.PendingBuildPath, j.ReadyBuildPath)
	if err != nil {
		return "", workErr("could not rename pending to ready path", err)
	}

	_, err = os.Lstat(j.LatestBuildPath)
	if err == nil {
		err = os.Remove(j.LatestBuildPath)
		if err != nil {
			return "", workErr("could not remove latest build link", err)
		}
	}

	err = os.Symlink(j.ReadyBuildPath, j.LatestBuildPath)
	if err != nil {
		return "", workErr("could not create latest build link", err)
	}

	return j.ReadyBuildPath, nil
}

// BootstrapProject bootstraps j's project if needed. This function is
// idempotent.
func BootstrapProject(j *Job) error {
	err := EnsureDirExists(j.RootBuildPath)
	if err != nil {
		return err
	}

	err = EnsureDirExists(filepath.Join(j.RootBuildPath, "pending"))
	if err != nil {
		return err
	}

	err = EnsureDirExists(filepath.Join(j.RootBuildPath, "ready"))
	if err != nil {
		return err
	}

	if j.Group != "" {
		err = EnsureDirExists(filepath.Join(j.RootBuildPath, "groups"))
		if err != nil {
			return err
		}
	}

	return nil
}

func workErr(s string, e error) error {
	s = "work: " + s
	if e != nil {
		s += "; " + e.Error()
	}
	return errors.New(s)
}

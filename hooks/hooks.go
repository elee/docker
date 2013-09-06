package hooks

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	//TO REMOVE
	"github.com/dotcloud/docker/utils"
)

var (
	registeredHooks = make(map[string][]*Hook)
)

type Hook struct {
	Name string // Filepath

	root     string // Root path
	fileName string
	hookName string
	prefix	 string
	action   string
}

func LoadAll(root, prefix string) error {
	err := filepath.Walk(root, func(path string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !fileInfo.IsDir() {
			p, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			if err := NewHook(root, p, prefix); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func NewHook(root, name, prefix string) error {
	if name[0] == '/' {
		name = name[1:]
	}
	h := &Hook{
		Name: name,
		root: root,
		prefix: prefix,
	}
	var action string
	parts := strings.Split(name, "/")
	h.hookName = parts[0]
	if len(parts) > 2 {
		action = parts[1]
	}
	h.action = action
	h.fileName = filepath.Base(name)

	hooks, exits := registeredHooks[h.hookName]
	if !exits {
		hooks = make([]*Hook, 0)
	}
	hooks = append(hooks, h)
	registeredHooks[h.hookName] = hooks
	//TO REMOVE
	utils.Debugf("Registering a new hook in %s/%s", h.hookName, action)
	return nil
}

func Execute(hook, action string, env []string) error {
	if hooks, exists := registeredHooks[hook]; exists {

		Sort(hooks)

		for _, h := range hooks {
			if h.action == "" || h.action == action {
				cmd := exec.Command(filepath.Join(h.root, h.Name))
				cmd.Env = append(env, fmt.Sprintf("%s_ACTION=%s_%s", h.prefix, hook, action))

				if err := cmd.Run(); err != nil {
					return fmt.Errorf("Hook failure: %s Error: %s", h.Name, err)
				}
			}
		}
	}
	return nil
}
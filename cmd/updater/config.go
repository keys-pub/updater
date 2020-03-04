package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/keys-pub/updater/util"
	"github.com/pkg/errors"
)

type cfg struct {
	appName string
	values  map[string]string
}

func newConfig(appName string) (*cfg, error) {
	if appName == "" {
		return nil, errors.Errorf("no app name")
	}
	c := &cfg{
		appName: appName,
	}
	if err := c.Load(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *cfg) GetUpdateAuto() (bool, bool) {
	auto := c.GetBool("auto")
	autoSet := c.GetBool("auto-set")
	return auto, autoSet
}

func (c *cfg) SetUpdateAuto(b bool) error {
	c.SetBool("auto", b)
	c.SetBool("auto-set", true)
	return c.Save()
}

func (c *cfg) GetUpdateAutoOverride() bool {
	return c.GetBool("auto-override")
}

func (c *cfg) SetUpdateAutoOverride(b bool) error {
	c.SetBool("auto-override", b)
	return c.Save()
}

func (c *cfg) GetInstallID() string {
	return ""
}

func (c *cfg) SetInstallID(installID string) error {
	return errors.Errorf("unsupported")
}

func (c *cfg) IsLastUpdateCheckTimeRecent(d time.Duration) bool {
	n := c.GetInt("last-check", 0)
	t := util.TimeFromMillis(util.TimeMs(n))
	recent := time.Since(t) < d
	return recent
}

func (c *cfg) SetLastUpdateCheckTime() {
	ts := util.TimeToMillis(time.Now())
	c.SetInt("last-check", int(ts))
	if err := c.Save(); err != nil {
		logger.Errorf("Error saving last update check time: %s", err)
	}
}

func (c *cfg) SetLastAppliedVersion(s string) error {
	c.Set("last-version", s)
	return c.Save()
}

func (c *cfg) GetLastAppliedVersion() string {
	return c.Get("last-version", "")
}

// AppName returns current app name.
func (c *cfg) AppName() string {
	return c.appName
}

// AppDir is where app related files are persisted.
func (c *cfg) AppDir() string {
	p, err := c.AppPath("", false)
	if err != nil {
		panic(err)
	}
	return p
}

// LogsDir is where logs are written.
func (c *cfg) LogsDir() string {
	p, err := c.LogsPath("", false)
	if err != nil {
		panic(err)
	}
	return p
}

// AppPath ...
func (c *cfg) AppPath(fileName string, makeDir bool) (string, error) {
	return SupportPath(c.AppName(), fileName, makeDir)
}

// LogsPath ...
func (c *cfg) LogsPath(fileName string, makeDir bool) (string, error) {
	return LogsPath(c.AppName(), fileName, makeDir)
}

// SupportPath ...
func SupportPath(appName string, fileName string, makeDir bool) (string, error) {
	switch runtime.GOOS {
	case "darwin":
		dir := filepath.Join(DefaultHomeDir(), "Library", "Application Support")
		return configPath(dir, appName, fileName, makeDir)
	case "windows":
		dir := os.Getenv("LOCALAPPDATA")
		if dir == "" {
			panic("LOCALAPPDATA not set")
		}
		return configPath(dir, appName, fileName, makeDir)
	case "linux":
		dir := os.Getenv("XDG_DATA_HOME")
		if dir == "" {
			dir = filepath.Join(DefaultHomeDir(), ".local", "share")
		}
		return configPath(dir, appName, fileName, makeDir)
	default:
		panic(fmt.Sprintf("unsupported platform %s", runtime.GOOS))
	}

}

// LogsPath ...
func LogsPath(appName string, fileName string, makeDir bool) (string, error) {
	switch runtime.GOOS {
	case "darwin":
		dir := filepath.Join(DefaultHomeDir(), "Library", "Logs")
		return configPath(dir, appName, fileName, makeDir)
	case "windows":
		dir := os.Getenv("LOCALAPPDATA")
		if dir == "" {
			panic("LOCALAPPDATA not set")
		}
		return configPath(dir, appName, fileName, makeDir)
	case "linux":
		dir := os.Getenv("XDG_CACHE_HOME")
		if dir == "" {
			dir = filepath.Join(DefaultHomeDir(), ".cache")
		}
		return configPath(dir, appName, fileName, makeDir)
	default:
		panic(fmt.Sprintf("unsupported platform %s", runtime.GOOS))
	}
}

func configPath(dir string, appName string, fileName string, makeDir bool) (string, error) {
	if appName == "" {
		return "", errors.Errorf("appName not specified")
	}
	dir = filepath.Join(dir, appName)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		logger.Infof("Creating directory: %s", dir)
		err := os.MkdirAll(dir, 0700)
		if err != nil {
			return "", err
		}
	}
	path := dir
	if fileName != "" {
		path = filepath.Join(path, fileName)
	}
	return path, nil
}

// DefaultHomeDir returns current user home directory (or "" on error).
func DefaultHomeDir() string {
	// TODO: Switch to UserHomeDir in go 1.12
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	return usr.HomeDir
}

func (c *cfg) Load() error {
	path, err := c.AppPath("config.json", false)
	if err != nil {
		return err
	}
	var values map[string]string
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		b, err := ioutil.ReadFile(path) // #nosec
		if err != nil {
			return err
		}
		if err := json.Unmarshal(b, &values); err != nil {
			return err
		}
	}
	if values == nil {
		values = map[string]string{}
	}
	c.values = values
	return nil
}

// Save ...
func (c *cfg) Save() error {
	path, err := c.AppPath("config.json", true)
	if err != nil {
		return err
	}
	b, err := json.Marshal(c.values)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(path, b, 0600); err != nil {
		return err
	}
	return nil
}

// Reset removes saved values.
func (c *cfg) Reset() error {
	path, err := c.AppPath("config.json", true)
	if err != nil {
		return err
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(path)
}

// Export ...
func (c cfg) Export() ([]byte, error) {
	return json.MarshalIndent(c.values, "", "  ")
}

// Get config value.
func (c *cfg) Get(key string, dflt string) string {
	v, ok := c.values[key]
	if !ok {
		return dflt
	}
	return v
}

// GetInt gets config value as int.
func (c *cfg) GetInt(key string, dflt int) int {
	v, ok := c.values[key]
	if !ok {
		return dflt
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		logger.Warningf("config value %s not an int", key)
		return 0
	}
	return n

}

// GetBool gets config value as bool.
func (c *cfg) GetBool(key string) bool {
	v, ok := c.values[key]
	if !ok {
		return false
	}
	b, _ := truthy(v)
	return b
}

// SetBool sets bool value for key.
func (c *cfg) SetBool(key string, b bool) {
	c.Set(key, truthyString(b))
}

// SetInt sets int value for key.
func (c *cfg) SetInt(key string, n int) {
	c.Set(key, strconv.Itoa(n))
}

// Set value.
func (c *cfg) Set(key string, value string) {
	c.values[key] = value
}

func truthy(s string) (bool, error) {
	s = strings.TrimSpace(s)
	switch s {
	case "1", "t", "true", "y", "yes":
		return true, nil
	case "0", "f", "false", "n", "no":
		return false, nil
	default:
		return false, errors.Errorf("invalid value: %s", s)
	}
}

func truthyString(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

package appmain

import (
	"fmt"
	"runtime/debug"

	"pulseshield/app/bootlog"
	"pulseshield/app/config"
	"pulseshield/app/core"
	"pulseshield/app/dpi"
	"pulseshield/app/state"
	"pulseshield/app/uiwin"
)

func Run(version string) error {
	defer func() {
		if r := recover(); r != nil {
			bootlog.Append(fmt.Sprintf("panic: %v", r))
			bootlog.Append(string(debug.Stack()))
		}
		bootlog.Append("exiting")
	}()

	bootlog.Append("starting version=" + version)
	cfg, err := config.LoadOrCreate()
	if err != nil {
		bootlog.Append("config error: " + err.Error())
		return err
	}

	st := state.New(cfg)
	coreBinary, _ := core.ResolveBinaryPath()
	coreManager := core.NewManager(core.NewAdapter(coreBinary))
	dpiRunner := &dpi.Runner{}

	bootlog.Append("starting windows native ui")
	if err := uiwin.Run(version, st, coreManager, dpiRunner); err != nil {
		bootlog.Append("ui error: " + err.Error())
		return fmt.Errorf("ui start failed: %w", err)
	}
	return nil
}

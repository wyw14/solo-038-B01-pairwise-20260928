package folder

import "fmt"

type Controller struct {
	state *State
}

func NewController(state *State) *Controller {
	return &Controller{state: state}
}

func (controller *Controller) Prepare(speed, cutLength float64) error {
	if !controller.state.Configure(speed, cutLength) {
		controller.state.SetReady(false, "invalid folder configuration")
		return fmt.Errorf("folder configuration is invalid")
	}
	controller.state.SetReady(true, "")
	return nil
}

func (controller *Controller) Stop(reason string) {
	controller.state.SetReady(false, reason)
}

func (controller *Controller) Finish() error {
	if !controller.state.CompleteLot() {
		return fmt.Errorf("folder cannot complete lot while not ready")
	}
	return nil
}

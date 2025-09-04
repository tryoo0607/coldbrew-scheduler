package clientgo

type Controller interface {
	Run(stopCh <-chan struct{})
}

type controllerWrapper struct {
	internal interface{ Run(stopCh <-chan struct{}) }
}

func (w *controllerWrapper) Run(stopCh <-chan struct{}) {
	w.internal.Run(stopCh)
}

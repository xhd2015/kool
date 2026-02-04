package engine

import (
	"github.com/xhd2015/arc-orm/engine"
)

var Engine = engine.Getter(get)

var engineImpl engine.Engine

func Init(impl engine.Engine) {
	engineImpl = impl
}

func get() engine.Engine {
	return engineImpl
}

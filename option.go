package process

type Option struct {
	Cache  map[int32]*Process
	Parent bool
	Mem    bool
	Cpu    bool
}

func NewOption() *Option {
	opt := &Option{
		Cpu:    true,
		Mem:    true,
		Parent: true,
	}
	return opt
}

type OptionFunc func(*Option)

func (opt *Option) Hit(pid int32) *Process {
	if opt.Cache == nil {
		return nil
	}

	if p, ok := opt.Cache[pid]; ok {
		return p
	}

	return nil
}

func Cache(v map[int32]*Process) OptionFunc {
	return func(option *Option) {
		option.Cache = v
	}
}

func Parent(v bool) OptionFunc {
	return func(option *Option) {
		option.Parent = v
	}
}

func Mem(v bool) OptionFunc {
	return func(option *Option) {
		option.Mem = v
	}
}

func Cpu(v bool) OptionFunc {
	return func(option *Option) {
		option.Cpu = v
	}
}

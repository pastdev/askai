package optional

import (
	"fmt"
	"strconv"

	"github.com/openai/openai-go/v3/packages/param"
	"github.com/spf13/pflag"
)

var _ pflag.Value = &optional[int64]{}

type optionalTypes interface {
	bool | float64 | int64
}

type optional[T optionalTypes] struct {
	value *param.Opt[T]
}

func valueOf[T optionalTypes](v string) (T, error) {
	casted := *new(T)
	switch any(casted).(type) {
	case bool:
		val, err := strconv.ParseBool(v)
		if err != nil {
			return casted, fmt.Errorf("parse bool: %w", err)
		}
		return any(val).(T), nil
	case float64:
		val, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return casted, fmt.Errorf("parse float: %w", err)
		}
		return any(val).(T), nil
	case int64:
		val, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return casted, fmt.Errorf("parse int: %w", err)
		}
		return any(val).(T), nil
	default:
		// should be impossible given type constraint, but we have here in case
		return casted, fmt.Errorf("unsupported type %T", casted)
	}
}

// IsBoolFlag implements [pflag.IsBoolFlag].
func (o *optional[T]) IsBoolFlag() bool {
	casted := *new(T)
	if _, ok := any(casted).(bool); ok {
		return true
	}
	return false
}

// Set implements [pflag.Value].
func (o *optional[T]) Set(v string) error {
	val, err := valueOf[T](v)
	if err != nil {
		return err
	}

	*o.value = param.NewOpt(val)
	return nil
}

// String implements [pflag.Value].
func (o *optional[T]) String() string {
	return o.value.String()
}

// Type implements [pflag.Value].
func (o *optional[T]) Type() string {
	return fmt.Sprintf("%T", *new(T))
}

func Var[T optionalTypes](
	f *pflag.FlagSet,
	p *param.Opt[T],
	name string,
	usage string,
) {
	VarP(f, p, name, "", usage)
}

func VarP[T optionalTypes](
	f *pflag.FlagSet,
	p *param.Opt[T],
	name string,
	shorthand string,
	usage string,
) {
	opt := &optional[T]{value: p}
	flag := f.VarPF(opt, name, shorthand, usage)
	// it seems like this shouldn't be necessary because the flag parsing is
	// supposed to call IsBoolFlag to determine if an arg is expected, but in
	// practice it appears the allowance for no arg is gated on the NoOptDefVal
	// property:
	//   https://github.com/spf13/pflag/blob/00153c6ac5d57e570f8378dedf321f34e373e4c3/flag.go#L1020
	//   https://github.com/spf13/pflag/blob/00153c6ac5d57e570f8378dedf321f34e373e4c3/flag.go#L1087
	if opt.IsBoolFlag() {
		flag.NoOptDefVal = "true"
	}
}

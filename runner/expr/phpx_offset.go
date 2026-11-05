package expr

import (
	"github.com/titpetric/phpx/internal/phpval"
	"github.com/titpetric/phpx/model"
)

// phpxOffsetCall runs an ArrayAccess operation through the runtime helper the
// environment carries beside __set.
func phpxOffsetCall(env *Env, op string, obj *model.Object, rest ...any) (any, error) {
	fn, ok := env.Base["__phpx_offset"].(func(...any) (any, error))
	if !ok {
		return nil, nil
	}
	return fn(append([]any{op, obj}, rest...)...)
}

// phpxQuietIndex compiles isset($object[$k]) and empty($object[$k]) over an
// ArrayAccess object: isset asks offsetExists alone and empty adds offsetGet
// when it answers true, as php does. Any other operand compiles as it does
// everywhere else, which the false in the second result says.
func (dc *directCompiler) phpxQuietIndex(call string, arg model.Expr) (closure, bool, error) {
	idx, ok := model.UnwrapParenthesized(arg).(*model.Index)
	if !ok || idx.Index == nil {
		return nil, false, nil
	}
	base, err := dc.compile(idx.Base)
	if err != nil {
		return nil, true, err
	}
	key, err := dc.compile(idx.Index)
	if err != nil {
		return nil, true, err
	}
	plain, err := dc.compile(arg)
	if err != nil {
		return nil, true, err
	}
	isEmpty := call == "empty" || call == "EMPTY"
	return func(env *Env) (any, error) {
		b, err := base(env)
		if err != nil {
			return nil, err
		}
		obj, isObj := b.(*model.Object)
		if str, isStr := b.(string); isStr {
			k, err := key(env)
			if err != nil {
				return nil, err
			}
			if !phpxStringOffsetSet(str, k) {
				return nil, nil
			}
			return plain(env)
		}
		if !isObj {
			return plain(env)
		}
		k, err := key(env)
		if err != nil {
			return nil, err
		}
		exists, err := phpxOffsetCall(env, "exists", obj, k)
		if err != nil || exists == nil || exists == false {
			return nil, err
		}
		if !isEmpty {
			return true, nil
		}
		return phpxOffsetCall(env, "get", obj, k)
	}, true, nil
}

// phpxStringOffsetSet reports whether isset($string[$k]) holds: an integer
// offset, or a string that spells one, inside the string counted from either end.
func phpxStringOffsetSet(s string, k any) bool {
	var i int64
	switch t := k.(type) {
	case int64:
		i = t
	case int:
		i = int64(t)
	case bool:
		if t {
			i = 1
		}
	case float64:
		i = int64(t)
	case string:
		n, ok := phpval.StringNumber(t)
		v, isInt := n.(int64)
		if !ok || !isInt {
			return false
		}
		i = v
	default:
		return false
	}
	if i < 0 {
		i += int64(len(s))
	}
	return i >= 0 && i < int64(len(s))
}

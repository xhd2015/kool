package flag

import "strings"

// ParseFlag parses a flag from the args slice.
//
// Usage example:
//
//	for i := 0; i < n; i++ {
//	  flag, value := flag.ParseFlag(args, &i)
//	  if flag == "" {
//	    remainArgs = append(remainArgs, args[i])
//	    continue
//	  }
//	  switch flag {
//	  case "-t", "--timeout":
//	    value, ok := value()
//	    if !ok {
//	      return fmt.Errorf("%s requires a value", flag)
//	    }
//	    ...
//	  default:
//	    return fmt.Errorf("unknown flag: %s", flag)
//	  }
//	}
func ParseFlag(args []string, i *int) (string, func() (string, bool)) {
	arg := args[*i]
	if !strings.HasPrefix(arg, "-") {
		return "", nil
	}
	idx := strings.Index(arg[1:], "=")

	flag := arg
	value := func() (string, bool) {
		if *i+1 >= len(args) {
			return "", false
		}
		*i++
		return args[*i], true
	}
	if idx >= 0 {
		idx++
		flag = arg[:idx]
		value = func() (string, bool) {
			return arg[idx+1:], true
		}
	}
	return flag, value
}

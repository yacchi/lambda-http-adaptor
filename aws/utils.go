package aws

func ExpandPathParameters(s string, parameters map[string]string) string {
	buf := make([]byte, 0, 2*len(s))
	for j := 0; j < len(s); j++ {
		if s[j] == '{' && j+1 < len(s) {
			name := ""
			for k := j + 1; k < len(s); k++ {
				if s[k] == '}' {
					name = s[j+1 : k]

					if name == "proxy+" {
						return string(buf) + parameters["proxy"]
					} else if v, ok := parameters[name]; ok {
						buf = append(buf, v...)
					} else {
						buf = append(buf, s[j:k+1]...)
					}
					j = k
					break
				}
			}
			if name == "" {
				return string(buf) + s[j:]
			}
		} else {
			buf = append(buf, s[j])
		}
	}
	return string(buf)
}

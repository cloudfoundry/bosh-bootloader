package terraform

type Outputs struct {
	Map map[string]interface{}
}

func (o Outputs) GetString(key string) string {
	if val, ok := o.Map[key]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

func (o Outputs) GetStringSlice(key string) []string {
	values := []string{}
	if _, ok := o.Map[key]; !ok {
		return values
	}

	if _, ok := o.Map[key].([]string); ok {
		return o.Map[key].([]string)
	}

	if _, ok := o.Map[key].([]interface{}); ok {
		for _, value := range o.Map[key].([]interface{}) {
			if _, ok := value.(string); !ok {
				return []string{}
			}
			values = append(values, value.(string))
		}
	}
	return values
}

func (o Outputs) GetStringMap(key string) map[string]string {
	if _, ok := o.Map[key].(map[string]string); ok {
		return o.Map[key].(map[string]string)
	}
	stringMap := map[string]string{}
	if _, ok := o.Map[key].(map[string]interface{}); ok {
		for k, v := range o.Map[key].(map[string]interface{}) {
			if _, ok := v.(string); !ok {
				return map[string]string{}
			}
			stringMap[k] = v.(string)
		}
	}
	return stringMap
}

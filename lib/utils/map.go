package utils

func Map[TI any, TO any](s []TI, fun func(TI) TO) []TO {
	outputSlice := make([]TO, len(s))
	for i, v := range s {
		outputSlice[i] = fun(v)
	}
	return outputSlice
}

package v2rayprobe

func Map[T, U any](input T, mapper func(T) U) U {
	return mapper(input)
}

func Map2[inp1, inp2, out any](input1 inp1, input2 inp2, mapper func(inp1, inp2) out) out {
	return mapper(input1, input2)
}
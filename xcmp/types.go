package xcmp

type (
	IntegerTypes interface {
		~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
	}

	FloatTypes interface {
		~float32 | ~float64
	}
)

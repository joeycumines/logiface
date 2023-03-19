package logiface

type (
	ArraySupport[E Event, A any] interface {
		NewArray() A

		AddArray(evt E, key string, arr A)

		AppendField(arr A, val any) A

		CanAppendArray() bool
		AppendArray(arr A, val A) A

		mustEmbedUnimplementedArraySupport()
	}

	// arraySupport is available via loggerShared.array, and models an external
	// array builder implementation.
	arraySupport[E Event] struct {
		iface       iArraySupport[E]
		newArray    func() any
		addArray    func(evt E, key string, arr any)
		appendField func(arr, val any) any
		appendArray func(arr, val any) any
	}

	// iArraySupport are the [ArraySupport] methods without array-specific behavior
	// (e.g. flags / checking if certain methods can be used)
	iArraySupport[E Event] interface {
		CanAppendArray() bool
	}

	UnimplementedArraySupport[E Event, A any] struct{}

	sliceArraySupport[E Event] struct{}
)

// WithArraySupport configures the implementation the logger uses to back the
// [Array] / [ArrayBuilder] implementation.
//
// By default, slices of type `[]any` are used.
func WithArraySupport[E Event, A any](impl ArraySupport[E, A]) Option[E] {
	return func(c *loggerConfig[E]) {
		if impl == nil {
			c.array = nil
		} else {
			c.array = newArraySupport(impl)
		}
	}
}

func newArraySupport[E Event, A any](impl ArraySupport[E, A]) *arraySupport[E] {
	return &arraySupport[E]{
		iface:    impl,
		newArray: func() any { return impl.NewArray() },
		addArray: func(evt E, key string, arr any) {
			impl.AddArray(evt, key, arr.(A))
		},
		appendField: func(arr, val any) any {
			return impl.AppendField(arr.(A), val)
		},
		appendArray: func(arr, val any) any {
			return impl.AppendArray(arr.(A), val.(A))
		},
	}
}

func generifyArraySupport[E Event](array *arraySupport[E]) *arraySupport[Event] {
	return &arraySupport[Event]{
		iface:    array.iface,
		newArray: array.newArray,
		addArray: func(evt Event, key string, arr any) {
			array.addArray(evt.(E), key, arr)
		},
		appendField: array.appendField,
		appendArray: array.appendArray,
	}
}

func (UnimplementedArraySupport[E, A]) CanAppendArray() bool { return false }

func (UnimplementedArraySupport[E, A]) AppendArray(arr A, val A) A {
	panic("not implemented")
}

func (UnimplementedArraySupport[E, A]) mustEmbedUnimplementedArraySupport() {}

func (x sliceArraySupport[E]) NewArray() []any { return nil }

func (x sliceArraySupport[E]) AddArray(evt E, key string, arr []any) {
	evt.AddField(key, arr)
}

func (x sliceArraySupport[E]) AppendField(arr []any, val any) []any {
	return append(arr, val)
}

func (x sliceArraySupport[E]) CanAppendArray() bool { return true }

func (x sliceArraySupport[E]) AppendArray(arr []any, val []any) []any {
	return append(arr, val)
}

func (x sliceArraySupport[E]) mustEmbedUnimplementedArraySupport() {}
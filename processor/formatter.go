package processor

import (
	"bytes"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/event"
	"github.com/codecrafters-io/redis-starter-go/spec"
)

type Formatter struct{}

func NewFormatter() *Formatter {
	return &Formatter{}
}

func (f *Formatter) FormatHandler() *formatHandler {
	return &formatHandler{
		formatter: f,
	}
}

func (f *Formatter) Format(data spec.Data) []byte {
	var bb bytes.Buffer

	f.formatTo(data, &bb)

	return bb.Bytes()
}

func (f *Formatter) formatTo(data spec.Data, bb *bytes.Buffer) {
	switch data.(type) {
	case *spec.SimpleStringData:
		bb.WriteByte(SimpleStringPrefix)
		str, _ := spec.Value[string](data)
		bb.WriteString(str)
		bb.WriteString("\r\n")

	case *spec.SimpleErrorData:
		bb.WriteByte(SimpleErrorPrefix)
		err, _ := spec.Value[error](data)
		bb.WriteString(err.Error())
		bb.WriteString("\r\n")

	case *spec.IntegerData:
		bb.WriteByte(IntegerPrefix)
		i, _ := spec.Value[int64](data)
		bb.WriteString(strconv.FormatInt(i, 10))
		bb.WriteString("\r\n")

	case *spec.BulkStringData:
		bb.WriteByte(BulkStringPrefix)

		str, _ := spec.Value[string](data)

		bb.WriteString(strconv.Itoa(len(str)))
		bb.WriteString("\r\n")

		bb.WriteString(str)
		bb.WriteString("\r\n")

	case *spec.ArrayData:
		bb.WriteByte(ArrayPrefix)
		bb.WriteString("\r\n")

		arr, _ := spec.Value[[]spec.Data](data)
		for _, d := range arr {
			f.formatTo(d, bb)
		}

		bb.WriteString("\r\n")
	}
}

var _ event.Handler = (*formatHandler)(nil)

type formatHandler struct {
	formatter *Formatter
}

func (f *formatHandler) Target() event.Type {
	return event.FormatEventType
}

func (f *formatHandler) Handle(e event.Event, push func(event.Event)) error {
	formatEvent, isType := e.(*event.FormatEvent)
	if !isType {
		return event.ErrInvalidEventType
	}

	b := f.formatter.Format(formatEvent.Data)
	push(&event.WriteEvent{
		ID_:  formatEvent.ID(),
		Data: b,
	})

	return nil
}

package tlc

import "math"

// State streams use the source IValueInput/OutputStream capabilities. Public
// byte-stream entry points and pool models share the same state mutation order.
type stateOutputStream interface {
	WriteShortNat(int16) error
	WriteLongNat(int64) error
	Write(Value) error
}

type stateInputStream interface {
	ReadShortNat() (int16, error)
	ReadLongNat() (int64, error)
	Read() (Value, error)
}

func (s *TLCStateMut) Write(out *ValueOutputStream) error {
	return s.writeStateStream(out)
}

func (s *TLCStateMut) writeStateStream(out stateOutputStream) error {
	if s == nil {
		panic(NewNullPointerException())
	}
	if s.functional {
		return NewWrongInvocationException("TLCStateFun.write: This is a TLC bug.")
	}
	if s.level > math.MaxInt16 {
		return NewTLCRuntimeException(ECTLCTraceTooLong, s.String())
	}
	if isNil(out) {
		panic(NewNullPointerException())
	}
	if err := out.WriteShortNat(s.WorkerID); err != nil {
		return err
	}
	if err := out.WriteLongNat(s.UID); err != nil {
		return err
	}
	if err := out.WriteShortNat(int16(s.level)); err != nil {
		return err
	}
	if s.printRecord != nil {
		return nil
	}
	for _, value := range s.values {
		if isNil(value) {
			panic(NewNullPointerException())
		}
		if err := out.Write(value); err != nil {
			return err
		}
	}
	return nil
}

func (s *TLCStateMut) Read(in *ValueInputStream) error {
	return s.readStateStream(in)
}

func (s *TLCStateMut) readStateStream(in stateInputStream) error {
	if s == nil {
		panic(NewNullPointerException())
	}
	if s.functional {
		return NewWrongInvocationException("TLCStateFun.read: This is a TLC bug.")
	}
	if isNil(in) {
		panic(NewNullPointerException())
	}
	workerID, err := in.ReadShortNat()
	if err != nil {
		return err
	}
	s.WorkerID = workerID
	uid, err := in.ReadLongNat()
	if err != nil {
		return err
	}
	s.UID = uid
	level, err := in.ReadShortNat()
	if err != nil {
		return err
	}
	s.level = int(level)
	if s.level < 0 {
		panic(NewAssertionError())
	}
	if s.printRecord != nil {
		return nil
	}
	for i := range s.values {
		value, err := in.Read()
		if err != nil {
			return err
		}
		s.values[i] = value
	}
	return nil
}

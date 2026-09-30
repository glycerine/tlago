package tlc

import "math"

func (s *TLCStateMut) Write(out *ValueOutputStream) error {
	if s == nil {
		return newTLCError(ECGeneral, "cannot write nil TLC state")
	}
	if s.level > math.MaxInt16 {
		return newTLCError(ECTLCTraceTooLong, "%s", s.String())
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
	for _, value := range s.values {
		if value == nil {
			return newTLCError(ECTLCStateNotCompletelySpecifiedNext, "%s", s.String())
		}
		if err := out.Write(value); err != nil {
			return err
		}
	}
	return nil
}

func (s *TLCStateMut) Read(in *ValueInputStream) error {
	workerID, err := in.ReadShortNat()
	if err != nil {
		return err
	}
	uid, err := in.ReadLongNat()
	if err != nil {
		return err
	}
	level, err := in.ReadShortNat()
	if err != nil {
		return err
	}
	s.WorkerID = workerID
	s.UID = uid
	s.level = int(level)
	if len(s.values) != len(stateVariables) {
		s.values = make([]Value, len(stateVariables))
	}
	for i := range s.values {
		value, err := in.Read()
		if err != nil {
			return err
		}
		s.values[i] = value
	}
	s.action = nil
	return nil
}

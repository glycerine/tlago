package tlc

func (s *TLCStateMut) Write(out *ValueOutputStream) error {
	if s == nil {
		if err := out.WriteInt(0); err != nil {
			return err
		}
		return nil
	}
	if err := out.WriteInt(int32(len(s.values))); err != nil {
		return err
	}
	for _, value := range s.values {
		assigned := value != nil
		if err := out.WriteBool(assigned); err != nil {
			return err
		}
		if assigned {
			if err := out.WriteExternal(value); err != nil {
				return err
			}
		}
	}
	if err := out.WriteInt(int32(s.level)); err != nil {
		return err
	}
	if s.action == nil {
		return out.WriteBool(false)
	}
	if err := out.WriteBool(true); err != nil {
		return err
	}
	return out.WriteUniqueString(UniqueStringOf(s.action.GetName()))
}

func (s *TLCStateMut) Read(in *ValueInputStream) error {
	count, err := in.ReadInt()
	if err != nil {
		return err
	}
	if int(count) != len(s.values) {
		s.values = make([]Value, int(count))
	}
	for i := 0; i < int(count); i++ {
		assigned, err := in.ReadBool()
		if err != nil {
			return err
		}
		if !assigned {
			s.values[i] = nil
			continue
		}
		value, err := in.ReadExternal()
		if err != nil {
			return err
		}
		s.values[i] = value
	}
	level, err := in.ReadInt()
	if err != nil {
		return err
	}
	s.level = int(level)
	hasAction, err := in.ReadBool()
	if err != nil {
		return err
	}
	if hasAction {
		name, err := in.readExternalUniqueString()
		if err != nil {
			return err
		}
		s.action = &Action{Name: name.String()}
	} else {
		s.action = nil
	}
	return nil
}

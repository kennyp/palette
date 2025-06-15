package types

import "unicode/utf16"

type UnicodeString string

func (s UnicodeString) MarshalBinary() (data []byte, err error) {
	runes := []rune(s)
	encoded := utf16.Encode(runes)

	data = make([]byte, len(encoded)*2)
	for i, u16 := range encoded {
		data[i*2] = byte(u16 >> 8)
		data[i*2+1] = byte(u16)
	}

	return data, nil
}

func (s UnicodeString) String() string {
	return string(s)
}

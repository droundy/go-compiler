package elf

import (
	"io"
	"os"
	"encoding/binary"
)

type Ptr uint32
const Page Ptr = 0x1000 // 4k

func (p Ptr) WriteTo(w io.Writer) os.Error {
	return binary.Write(w, binary.LittleEndian, uint32(p))
}

// We only create a single text segment and a single data segment.
// This is a little sloppy, but simplifies things.

type Header struct {
	Entry Ptr
	Text []byte
	Data []byte
}

type Section struct {
}

func (h *Header) WriteTo(w io.Writer) (err os.Error) {
	// e_ident is sixteen bytes
	_,err = w.Write([]byte("\x7fELF\001\001\001\000\000\000\000\000\000\000\000\000"))
	if err != nil { return }
	// e_type = ET_EXEC
	if err = binary.Write(w, binary.LittleEndian, uint16(2)); err != nil {
		return
	}
	// e_machine = EM_386
	if err = binary.Write(w, binary.LittleEndian, uint16(3)); err != nil {
		return
	}
	// e_version = EV_CURRENT
	if err = binary.Write(w, binary.LittleEndian, uint32(1)); err != nil {
		return
	}
	// e_entry = h.Entry
	if err = h.Entry.WriteTo(w); err != nil {
		return
	}
	var off Ptr = 36+16
	// e_phoff = 36 + 16 (which is the size of the header...
	if err = off.WriteTo(w); err != nil {
		return
	}
	// e_shoff = 0 // no section headers!
	if err = Ptr(0).WriteTo(w); err != nil {
		return
	}
	// e_flags = 0 (correct for x86)
	if err = Ptr(0).WriteTo(w); err != nil {
		return
	}
	// e_ehsize = off (right now)
	if err = binary.Write(w, binary.LittleEndian, uint16(off)); err != nil {
		return
	}
	// e_phentsize = 8*4
	if err = binary.Write(w, binary.LittleEndian, uint16(8*4)); err != nil {
		return
	}
	// e_phentnum = 2
	if err = binary.Write(w, binary.LittleEndian, uint16(2)); err != nil {
		return
	}
	// e_shentsize = 0 (unused)
	if err = binary.Write(w, binary.LittleEndian, uint16(0)); err != nil {
		return
	}
	// e_shentnum = 0
	if err = binary.Write(w, binary.LittleEndian, uint16(0)); err != nil {
		return
	}
	// e_shstrndx = SHN_UNDEF = 0
	if err = binary.Write(w, binary.LittleEndian, uint16(0)); err != nil {
		return
	}
	return nil
}

func WriteProgramHeaderText(w io.Writer, offset Ptr, vaddr Ptr, text []byte) (err os.Error) {
	// p_type = PT_LOAD
	if err = binary.Write(w, binary.LittleEndian, uint32(1)); err != nil {
		return
	}
	// p_offset = offset
	if err = offset.WriteTo(w); err != nil {
		return
	}
	// p_vaddr = vaddr
	if err = vaddr.WriteTo(w); err != nil {
		return
	}
	// p_paddr = unspecified (I'll just use zero)
	if err = Ptr(0).WriteTo(w); err != nil {
		return
	}
	// p_filesz = len(text)
	if err = Ptr(len(text)).WriteTo(w); err != nil {
		return
	}
	// p_memsz = len(text)
	if err = Ptr(len(text)).WriteTo(w); err != nil {
		return
	}
	// p_flags = 5 (PF_X = 1, PF_R = 4)
	if err = Ptr(5).WriteTo(w); err != nil {
		return
	}
	// p_align = alignment of vaddr and offset, we'll find this
	// automagically...
	for i:=Page; i>=0; i >>= 1 {
		if offset % i == vaddr % i {
			if err = i.WriteTo(w); err != nil {
				return
			}
			break
		}
	}
	return
}

func WriteProgramHeaderData(w io.Writer, offset Ptr, vaddr Ptr, data []byte) (err os.Error) {
	// p_type = PT_LOAD
	if err = binary.Write(w, binary.LittleEndian, uint32(1)); err != nil {
		return
	}
	// p_offset = offset
	if err = offset.WriteTo(w); err != nil {
		return
	}
	// p_vaddr = vaddr
	if err = vaddr.WriteTo(w); err != nil {
		return
	}
	// p_paddr = unspecified (I'll just use zero)
	if err = Ptr(0).WriteTo(w); err != nil {
		return
	}
	// p_filesz = len(data)
	if err = Ptr(len(data)).WriteTo(w); err != nil {
		return
	}
	// p_memsz = len(data)
	if err = Ptr(len(data)).WriteTo(w); err != nil {
		return
	}
	// p_flags = 7 (PF_X = 1, PF_W = 2, PF_R = 4)
	if err = Ptr(7).WriteTo(w); err != nil {
		return
	}
	// p_align = alignment of vaddr and offset, we'll find this
	// automagically...
	for i:=Page; i>=0; i >>= 1 {
		if offset % i == vaddr % i {
			if err = i.WriteTo(w); err != nil {
				return
			}
			break
		}
	}
	return
}

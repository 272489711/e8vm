package asm8

import (
	"strconv"

	"e8vm.io/e8vm/arch8"
	asminst "e8vm.io/e8vm/asm8/inst"
	"e8vm.io/e8vm/lex8"
)

var (
	// op reg reg imm(signed)
	opImsMap = map[string]uint32{
		"addi": arch8.ADDI,
		"slti": arch8.SLTI,
	}

	opMemMap = map[string]uint32{
		"lw":  arch8.LW,
		"lb":  arch8.LB,
		"lbu": arch8.LBU,
		"sw":  arch8.SW,
		"sb":  arch8.SB,
	}

	// op reg reg imm(unsigned)
	opImuMap = map[string]uint32{
		"andi": arch8.ANDI,
		"ori":  arch8.ORI,
		"xori": arch8.XORI,
	}

	// op reg imm(signed or unsigned)
	opImmMap = map[string]uint32{
		"lui": arch8.LUI,
	}
)

// parseImu parses an unsigned 16-bit immediate
func parseImu(p lex8.Logger, op *lex8.Token) uint32 {
	ret, e := strconv.ParseUint(op.Lit, 0, 32)
	if e != nil {
		p.Errorf(op.Pos, "invalid unsigned immediate %q: %s", op.Lit, e)
		return 0
	}

	if (ret & 0xffff) != ret {
		p.Errorf(op.Pos, "immediate too large: %s", op.Lit)
		return 0
	}

	return uint32(ret)
}

// parseIms parses an unsigned 16-bit immediate
func parseIms(p lex8.Logger, op *lex8.Token) uint32 {
	ret, e := strconv.ParseInt(op.Lit, 0, 32)
	if e != nil {
		p.Errorf(op.Pos, "invalid signed immediate %q: %s", op.Lit, e)
		return 0
	}

	if ret > 0x7fff || ret < -0x8000 {
		p.Errorf(op.Pos, "immediate out of 16-bit range: %s", op.Lit)
		return 0
	}

	return uint32(ret) & 0xffff
}

// parseImm parses an unsigned 16-bit immediate
func parseImm(p lex8.Logger, op *lex8.Token) uint32 {
	ret, e := strconv.ParseInt(op.Lit, 0, 32)
	if e != nil {
		p.Errorf(op.Pos, "invalid signed immediate %q: %s", op.Lit, e)
		return 0
	}

	if ret > 0xffff || ret < -0x8000 {
		p.Errorf(op.Pos, "immediate out of 16-bit range: %s", op.Lit)
		return 0
	}

	return uint32(ret) & 0xffff
}

func makeInstImm(op, d, s, im uint32) *inst {
	ret := asminst.Imm(op, d, s, im)
	return &inst{inst: ret}
}

func resolveInstImm(p lex8.Logger, ops []*lex8.Token) (*inst, bool) {
	op0 := ops[0]
	opName := op0.Lit
	args := ops[1:]

	var (
		op, d, s, im uint32
		pack, sym    string
		fill         int
		symTok       *lex8.Token
	)

	argCount := func(n int) bool {
		if !argCount(p, ops, n) {
			return false
		}
		if n >= 1 {
			d = resolveReg(p, args[0])
		}
		return true
	}

	parseSym := func(t *lex8.Token, f func(lex8.Logger, *lex8.Token) uint32) {
		if mightBeSymbol(t.Lit) {
			pack, sym = parseSym(p, t)
			fill = fillLow
			symTok = t
		} else {
			im = f(p, t)
		}
	}

	var found bool
	if op, found = opImsMap[opName]; found {
		// op reg reg imm(signed)
		if argCount(3) {
			s = resolveReg(p, args[1])
			parseSym(args[2], parseIms)
		}
	} else if op, found = opMemMap[opName]; found {
		if len(args) == 2 {
			// mem op can omit the offset if it is 0
			d = resolveReg(p, args[0])
			s = resolveReg(p, args[1])
		} else if argCount(3) {
			s = resolveReg(p, args[1])
			parseSym(args[2], parseIms)
		}
	} else if op, found = opImuMap[opName]; found {
		// op reg reg imm(unsigned)
		if argCount(3) {
			s = resolveReg(p, args[1])
			parseSym(args[2], parseImu)
		}
	} else if op, found = opImmMap[opName]; found {
		// op reg imm(signed or unsigned)
		if argCount(2) {
			parseSym(args[1], parseImm)
		}
		if opName == "lui" && fill == fillLow {
			fill = fillHigh
		}
	} else {
		return nil, false
	}

	ret := makeInstImm(op, d, s, im)
	ret.pkg = pack
	ret.sym = sym
	ret.fill = fill
	ret.symTok = symTok

	return ret, true
}

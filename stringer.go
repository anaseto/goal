// Code generated by "stringer -type=Monad,Dyad,Adverb,TokenType,ppTokenType,ppBlockType -output types_strings.go"; DO NOT EDIT.

package main

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[VReturn-0]
	_ = x[VFlip-1]
	_ = x[VNegate-2]
	_ = x[VFirst-3]
	_ = x[VClassify-4]
	_ = x[VEnum-5]
	_ = x[VWhere-6]
	_ = x[VReverse-7]
	_ = x[VAscend-8]
	_ = x[VDescend-9]
	_ = x[VGroup-10]
	_ = x[VNot-11]
	_ = x[VEnlist-12]
	_ = x[VSort-13]
	_ = x[VLen-14]
	_ = x[VFloor-15]
	_ = x[VString-16]
	_ = x[VNub-17]
	_ = x[VType-18]
	_ = x[VValues-19]
}

const _Monad_name = "VReturnVFlipVNegateVFirstVClassifyVEnumVWhereVReverseVAscendVDescendVGroupVNotVEnlistVSortVLenVFloorVStringVNubVTypeVValues"

var _Monad_index = [...]uint8{0, 7, 12, 19, 25, 34, 39, 45, 53, 60, 68, 74, 78, 85, 90, 94, 100, 107, 111, 116, 123}

func (i Monad) String() string {
	if i < 0 || i >= Monad(len(_Monad_index)-1) {
		return "Monad(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Monad_name[_Monad_index[i]:_Monad_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[VRight-0]
	_ = x[VAdd-1]
	_ = x[VSubtract-2]
	_ = x[VMultiply-3]
	_ = x[VDivide-4]
	_ = x[VMod-5]
	_ = x[VAnd-6]
	_ = x[VOr-7]
	_ = x[VLess-8]
	_ = x[VMore-9]
	_ = x[VEqual-10]
	_ = x[VMatch-11]
	_ = x[VConcat-12]
	_ = x[VCut-13]
	_ = x[VTake-14]
	_ = x[VDrop-15]
	_ = x[VCast-16]
	_ = x[VFind-17]
	_ = x[VApply-18]
	_ = x[VApplyN-19]
}

const _Dyad_name = "VRightVAddVSubtractVMultiplyVDivideVModVAndVOrVLessVMoreVEqualVMatchVConcatVCutVTakeVDropVCastVFindVApplyVApplyN"

var _Dyad_index = [...]uint8{0, 6, 10, 19, 28, 35, 39, 43, 46, 51, 56, 62, 68, 75, 79, 84, 89, 94, 99, 105, 112}

func (i Dyad) String() string {
	if i < 0 || i >= Dyad(len(_Dyad_index)-1) {
		return "Dyad(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Dyad_name[_Dyad_index[i]:_Dyad_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[AEach-0]
	_ = x[AFold-1]
	_ = x[AScan-2]
}

const _Adverb_name = "AEachAFoldAScan"

var _Adverb_index = [...]uint8{0, 5, 10, 15}

func (i Adverb) String() string {
	if i < 0 || i >= Adverb(len(_Adverb_index)-1) {
		return "Adverb(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Adverb_name[_Adverb_index[i]:_Adverb_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[EOF-0]
	_ = x[ERROR-1]
	_ = x[ADVERB-2]
	_ = x[IDENT-3]
	_ = x[LEFTBRACE-4]
	_ = x[LEFTBRACKET-5]
	_ = x[LEFTPAREN-6]
	_ = x[NEWLINE-7]
	_ = x[NUMBER-8]
	_ = x[RIGHTBRACE-9]
	_ = x[RIGHTBRACKET-10]
	_ = x[RIGHTPAREN-11]
	_ = x[SEMICOLON-12]
	_ = x[STRING-13]
	_ = x[VERB-14]
}

const _TokenType_name = "EOFERRORADVERBIDENTLEFTBRACELEFTBRACKETLEFTPARENNEWLINENUMBERRIGHTBRACERIGHTBRACKETRIGHTPARENSEMICOLONSTRINGVERB"

var _TokenType_index = [...]uint8{0, 3, 8, 14, 19, 28, 39, 48, 55, 61, 71, 83, 93, 102, 108, 112}

func (i TokenType) String() string {
	if i < 0 || i >= TokenType(len(_TokenType_index)-1) {
		return "TokenType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _TokenType_name[_TokenType_index[i]:_TokenType_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[ppSEP-0]
	_ = x[ppEOF-1]
	_ = x[ppCLOSE-2]
	_ = x[ppADVERB-3]
	_ = x[ppIDENT-4]
	_ = x[ppNUMBER-5]
	_ = x[ppSTRING-6]
	_ = x[ppVERB-7]
}

const _ppTokenType_name = "ppSEPppEOFppCLOSEppADVERBppIDENTppNUMBERppSTRINGppVERB"

var _ppTokenType_index = [...]uint8{0, 5, 10, 17, 25, 32, 40, 48, 54}

func (i ppTokenType) String() string {
	if i < 0 || i >= ppTokenType(len(_ppTokenType_index)-1) {
		return "ppTokenType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _ppTokenType_name[_ppTokenType_index[i]:_ppTokenType_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[ppBRACE-0]
	_ = x[ppBRACKET-1]
	_ = x[ppPAREN-2]
}

const _ppBlockType_name = "ppBRACEppBRACKETppPAREN"

var _ppBlockType_index = [...]uint8{0, 7, 16, 23}

func (i ppBlockType) String() string {
	if i < 0 || i >= ppBlockType(len(_ppBlockType_index)-1) {
		return "ppBlockType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _ppBlockType_name[_ppBlockType_index[i]:_ppBlockType_index[i+1]]
}

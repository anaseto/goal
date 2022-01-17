# atom types nil, int64, float64, string, blob ([]byte) (SQLite)
# list types []int64, []float64, []string, []Value
# no operator precedence rules (BQN)
1×(2+3) ≡ 1×2+3
# list auto-flattening? (Perl, maybe not, unclear)
# syntax: ; , {} [] ← ↩ () : "" ''
# compiles to Go
# Primitives:
# Glyph	Monadic			Dyadic
# +			flip?				Add
# - 		Negate			Subtract
# × 		Sign				Multiply
# ÷ 		Reciprocal	Divide
# ⌊ 		Floor				Minimum
# ⌈ 		Ceiling			Maximum
# ∧ 		Sort Up			And*
# ∨ 		Sort Down		Or*
# ¬ 		Not*				Span*					use ~ ?
# | 		Absolute		Modulus
# ≤									Less or Equal
# < 		Sort Up			Less Than
# > 		Sort Down		Greater Than
# ≥									Greater Than or Equal to
# = 	  Classify		Equals
# #			Length?			Replicate?
# ≠ 		Length?			Not Equals
# ≡ 		XXXXXX			Match
# ≢ 		XXXXXX			Not Match
# ⥊ 		Deshape			Reshape*
# ∾ 		Join*				Join to
# ⋈ 		Enlist*			Pair*
# ↑ 		First				Take
# ↓ 		Tail				Drop
# ↕ 		Range				Windows*
# » 		Nudge*			Shift Before*
# « 		Nudge Back*	Shift After*
# ⌽ 		Reverse			Rotate
# / 		Indices			Replicate
# ⍋ 		Grade Up		Bins Up
# ⍒ 		Grade Down	Bins Down
# ∊ 		Mark Firsts				Member of
# ⍷ 		Occurrence Count*	Find
# ? 		Deduplicate				Index Of
# ⊔ 		Group Indices*		Group*
# *,!		Indices Inv				Progressive Index of*
#
#
# Removed:
# ⋆ 		Exponential	Power
# √ 		Square Root	Root
# ≍ 		Solo*				Couple*
# ⍉ 		Transpose*	Reorder axes*
# ⊏ 		First Cell*	Select*					([] syntax for select)
# ⍷ 		Deduplicate				Find
# ⊑ 		First				Pick*
# ⊣ 		Identity		Left
# ⊢ 		Identity		Right
# ⊒ 		Occurrence Count*	Progressive Index of*

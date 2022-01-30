# atom types nil, int64, float64, string, blob ([]byte) (SQLite)
# list types []int64, []float64, []string, []Value
# no operator precedence rules (BQN)
1×(2+3) ≡ 1×2+3
# list auto-flattening? (Perl, maybe not, unclear)
# syntax: ; , {} [] ← ↩ () : "" ''
# compiles to Go
# Primitives:
# Glyph	Monadic			Dyadic
# +			flip				Add
# -			Negate			Subtract
# ×			Sign				Multiply
# ÷			Reciprocal	Divide
# ⌊			Floor				Minimum
# ⌈			Ceiling			Maximum
# ∧			XXXXXXXXX		And*
# ∨			XXXXXXXXX		Or*
# ¬			Not*				Span*
# |			Absolute		Modulus
# ≤									Less or Equal
# <			Sort Up			Less Than
# >			Sort Down		Greater Than
# ≥									Greater Than or Equal to
# =		  Classify?		Equals
# ≠			Length?			Not Equals
# ≡			XXXXXX			Match
# ≢			XXXXXX			Not Match
# ∾			Join*				Join to  (use ~ ?)
# ⋈			Enlist*			Pair*
# ↑			First				Take
# ↓			Tail				Drop
# ↕			Range				Windows*
# »			Nudge*			Shift Before*
# «			Nudge Back*	Shift After*
# ⌽			Reverse			Rotate
# /			Indices			Replicate
# \			Indices Inv	Expand (APL)
# ⍋			Grade Up		Bins Up
# ⍒			Grade Down	Bins Down
# ∊			Mark Firsts				Member of
# ⍷			Deduplicate				Find
# ⊒			Occurrence Count*	Progressive Index of*
# ⊐			Classify				  Index Of
# ⊔			Group Indices*		Group*
# !			Error							Annotate?
#
# Ideas:
# ? 		Deduplicate				Index Of
# #			Length?			Replicate?
# %			comment?
#
# Removed:
# ⋆ 		Exponential	Power
# √ 		Square Root	Root
# ≍ 		Solo*				Couple*
# ⍉ 		Transpose*	Reorder axes*
# ⊏ 		First Cell*	Select*					([] syntax for select)
# ⊑ 		First				Pick*
# ⊣ 		Identity		Left
# ⊢ 		Identity		Right
# ⥊ 		Deshape			Reshape*
#
# TODO:
# Optimization: avoid unecessary cloning using in-place operations

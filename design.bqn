# atom types nil, int64, float64, string, blob ([]byte) (SQLite)
# list types []int64, []float64, []string, []Value
# no operator precedence rules (BQN)
1×(2+3) ≡ 1×2+3
# list auto-flattening? (Perl, maybe not, unclear)
# compiles to Go
# Example:
Add_fact ← {𝕊⟨question,answer,notes,type,tags⟩
    ⟨question, answer, notes, type⟩ db.Eval "INSERT INTO facts(question, answer, notes, type) VALUES(?, ?, ?, ?)"
		fact_uid ← db.last_insert_rowid
		{"oneside": Add_card fact_uid;
		 "twoside": Add_card¨ ⟨fact_uid,fact_data⟩;
		 "cloze": 
				Process ← {
					cmd ← 1↓𝕩
					Check_Cloze_End cmd
					fact_data ← 𝕨 ∾ 1↓cmd
					Add_card ⟨fact_uid,fact_data⟩
				}
			  clozes ← Get_clozes question
				(↕≠clozes) Process clozes
			_ : ! "invalid type: "∾type
		}
		tags ∾←all
		# ....
    fact_uid
}
